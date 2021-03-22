package bootstrap

import (
	"context"
	"go.uber.org/fx"
)

var cliRunnerModule = &Module{
	Name: "CLI Runner",
	Precedence: CommandLineRunnerPrecedence,
	Options: []fx.Option{
		fx.Invoke(shutdown),
	},
}

// EnableCliRunnerMode should be called before Execute(), otherwise it won't run.
// The provided "runnerFuncs" are standard FX lifecycle functions that typically used with fx.Invoke(...)
//
// Typical CliRunner function looks like this:
//		func cliRunner(lc fx.Lifecycle, di OtherDependencies) {
//			bootstrap.RegisterCliRunnerHook(lc, func(ctx context.Context) error {
//				// Do your stuff
//				return err
//			})
//		}
//
// Using this pattern garuantees following things:
// 		1. The application is automatically shutdown after all lifecycle hooks finished
//		2. The runner funcs are run after all other fx.Invoke
// 		3. All other "OnStop" are executed relardless if any hook function returns error (graceful shutdown)
// 		3. If any hook functions returns error, it reflected as non-zero process exit code
// Note: calling this function repeatly would override previous invocation (i.e. only the last invocation takes effect)
func EnableCliRunnerMode(runnerFuncs ...interface{}) {
	cliRunnerModule.Options = append([]fx.Option{fx.Invoke(runnerFuncs...)}, fx.Invoke(shutdown))
	Register(cliRunnerModule)
}

func RegisterCliRunnerHook(lc fx.Lifecycle, fc func(ctx context.Context) error) {
	var err error
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err = fc(ctx)
			return nil
		},
		OnStop:  func(ctx context.Context) error {
			return err
		},
	})
}

func shutdown(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return shutdowner.Shutdown()
		},
	})
}

