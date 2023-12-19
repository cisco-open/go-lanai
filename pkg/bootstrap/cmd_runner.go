package bootstrap

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"go.uber.org/fx"
)

const (
	FxCliRunnerGroup    = "bootstrap_cli_runner"
	CliRunnerModuleName = "CLI Runner"
)

type CliRunner func(ctx context.Context) error

type CliRunnerEnabler interface {
	// EnableCliRunnerMode see bootstrap.EnableCliRunnerMode
	EnableCliRunnerMode(runnerProviders ...interface{})
}

// CliRunnerLifecycleHooks provide instrumentation around CliRunners
type CliRunnerLifecycleHooks interface {
	Before(ctx context.Context, runner CliRunner) context.Context
	After(ctx context.Context, runner CliRunner, err error) context.Context
}

// EnableCliRunnerMode should be called before Execute(), otherwise it won't run.
// "runnerProviders" are standard FX lifecycle functions that typically used with fx.Provide(...)
// signigure of "runnerProviders", but it should returns CliRunner, otherwise it won't run
//
// example runner provider:
//
//	func myRunner(di OtherDependencies) CliRunner {
//		return func(ctx context.Context) error {
//			// Do your stuff
//			return err
//		}
//	}
//
// Using this pattern guarantees following things:
//  1. The application is automatically shutdown after all lifecycle hooks finished
//  2. The runner funcs are run after all other fx.Invoke
//  3. All other "OnStop" are executed regardless if any hook function returns error (graceful shutdown)
//  4. If any hook functions returns error, it reflected as non-zero process exit code
//  5. Each cli runner are separately traced if tracing is enabled
//
// Note: calling this function repeatedly would override previous invocation (i.e. only the last invocation takes effect)
func EnableCliRunnerMode(runnerProviders ...interface{}) {
	enableCliRunnerMode(bootstrapper(), runnerProviders)
}

func newCliRunnerModule() *Module {
	return &Module{
		Name:       CliRunnerModuleName,
		Precedence: CommandLineRunnerPrecedence,
		Options:    []fx.Option{fx.Invoke(cliRunnerExec)},
	}
}

func enableCliRunnerMode(b *Bootstrapper, runnerProviders []interface{}) {
	// first find existing runner module or register one
	var cliRunnerModule *Module
LOOP:
	for v := range b.modules {
		switch m, ok := v.(*Module); {
		case ok && m != nil && m.Name == CliRunnerModuleName:
			cliRunnerModule = m
			break LOOP
		}
	}
	if cliRunnerModule == nil {
		cliRunnerModule = newCliRunnerModule()
		b.Register(cliRunnerModule)
	}

	// create annotated providers and add to module
	providers := make([]interface{}, len(runnerProviders))
	for i, provider := range runnerProviders {
		providers[i] = fx.Annotated{
			Group:  FxCliRunnerGroup,
			Target: provider,
		}
	}
	cliRunnerModule.Options = append(cliRunnerModule.Options, fx.Provide(providers...))
}

type cliDI struct {
	fx.In
	Hooks   []CliRunnerLifecycleHooks `group:"bootstrap_cli_runner"`
	Runners []CliRunner               `group:"bootstrap_cli_runner"`
}

func cliRunnerExec(lc fx.Lifecycle, shutdowner fx.Shutdowner, di cliDI) {
	order.SortStable(di.Hooks, order.OrderedFirstCompare)
	var err error
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for _, runner := range di.Runners {
				c := ctx
				// before hook
				for _, before := range di.Hooks {
					c = before.Before(c, runner)
				}
				// run
				err = runner(c)

				// after hook
				for _, after := range di.Hooks {
					c = after.After(c, runner, err)
				}
				if err != nil {
					break
				}
			}

			// we delay error reporting to OnStop
			return shutdowner.Shutdown()
		},
		OnStop: func(ctx context.Context) error {
			return err
		},
	})
}
