// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"go.uber.org/fx"
)

const (
	FxCliRunnerGroup    = "bootstrap_cli_runner"
	CliRunnerModuleName = "CLI Runner"
)

type CliRunner func(ctx context.Context) error

func (r CliRunner) WithOrder(order int) OrderedCliRunner {
	return OrderedCliRunner{
		Precedence: order,
		CliRunner:  r,
	}
}

type OrderedCliRunner struct {
	Precedence int
	CliRunner CliRunner
}

func (r OrderedCliRunner) Order() int {
	return r.Precedence
}

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
// signigure of "runnerProviders", but it should returns CliRunner or OrderedCliRunner, otherwise it won't run
//
// example of runner provider:
//
//	func myRunner(di OtherDependencies) CliRunner {
//		return func(ctx context.Context) error {
//			// Do your stuff
//			return err
//		}
//	}
//
// example of ordered runner provider:
//
//	func myRunner(di OtherDependencies) OrderedCliRunner {
//		return bootstrap.OrderedCliRunner{
//			Precedence: 0,
//			CliRunner:  func(ctx context.Context) error {
//				// Do your stuff
//				return err
//			},
//		}
//	}
//
// Using this pattern guarantees following things:
//  1. The application is automatically shutdown after all lifecycle hooks finished
//  2. The runner funcs are run after all other fx.Invoke
//  3. All other "OnStop" are executed regardless if any hook function returns error (graceful shutdown)
//  4. If any hook functions returns error, it reflected as non-zero process exit code
//  5. Each cli runner are separately traced if tracing is enabled
//  6. Any CliRunner without order is considered as having order 0
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
	for m := range b.modules {
		if m != nil && m.Name == CliRunnerModuleName {
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
	Hooks          []CliRunnerLifecycleHooks `group:"bootstrap_cli_runner"`
	Runners        []CliRunner               `group:"bootstrap_cli_runner"`
	OrderedRunners []OrderedCliRunner        `group:"bootstrap_cli_runner"`
}

func cliRunnerExec(lc fx.Lifecycle, shutdowner fx.Shutdowner, di cliDI) {
	order.SortStable(di.Hooks, order.OrderedFirstCompare)
	runners := make([]OrderedCliRunner, len(di.Runners), len(di.Runners)+len(di.OrderedRunners))
	for i := range di.Runners {
		runners[i] = di.Runners[i].WithOrder(0)
	}
	for i := range di.OrderedRunners {
		runners = append(runners, di.OrderedRunners[i])
	}
	order.SortStable(runners, order.OrderedFirstCompare)
	var err error
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for _, runner := range runners {
				c := ctx
				// before hook
				for _, before := range di.Hooks {
					c = before.Before(c, runner.CliRunner)
				}
				// run
				err = runner.CliRunner(c)

				// after hook
				for _, after := range di.Hooks {
					c = after.After(c, runner.CliRunner, err)
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
