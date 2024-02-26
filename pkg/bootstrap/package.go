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
	"github.com/cisco-open/go-lanai/pkg/log"
	"go.uber.org/fx"
	"time"
)

var logger = log.New("Bootstrap")

// InitModule returns the module that would run with highest priority
func InitModule(cliCtx *CliExecContext, app *App) *Module {
	return &Module{
		Precedence: HighestPrecedence,
		PriorityOptions: []fx.Option{
			fx.WithLogger(provideFxLogger),
			fx.Supply(cliCtx),
			fx.Supply(app),
			fx.Provide(provideApplicationContext),
			fx.Provide(provideBuildInfoResolver),
			fx.Invoke(bootstrap),
		},
	}
}

// MiscModules returns the module that would run with various precedence
func MiscModules() []*Module {
	return []*Module{
		{
			Precedence: StartupSummaryPrecedence,
			Options: []fx.Option{
				fx.Invoke(startupTiming), // startup need to be run at last
			},
		},
		{
			Precedence: HighestPrecedence,
			PriorityOptions: []fx.Option{
				// shutdown timing need to be run at last
				// note that fx.Hook.OnStop is run in reversed order
				fx.Invoke(shutdownTiming),
			},
		},
	}
}

func provideApplicationContext(app *App, config ApplicationConfig) *ApplicationContext {
	app.ctx.config = config
	return app.ctx
}

func provideBuildInfoResolver(appCtx *ApplicationContext) BuildInfoResolver {
	return newDefaultBuildInfoResolver(appCtx)
}

func bootstrap(lc fx.Lifecycle, ac *ApplicationContext) {
	logProperties := &log.Properties{}
	err := ac.config.Bind(logProperties, "log")
	if err == nil {
		err = log.UpdateLoggingConfiguration(logProperties)
	}
	if err != nil {
		logger.Error("Error updating logging configuration", "error", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.WithContext(ac).Info("On Application Start") //nolint:contextcheck
			return nil
		},
	})
}

func startupTiming(lc fx.Lifecycle, appCtx *ApplicationContext) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if t, ok := ctx.Value(ctxKeyStartTime).(time.Time); ok {
				elapsed := time.Now().Sub(t).Truncate(time.Millisecond)
				logger.WithContext(ctx).Infof("Started %s in %v", appCtx.Name(), elapsed)
			}
			return nil
		},
	})
}

func shutdownTiming(lc fx.Lifecycle, appCtx *ApplicationContext) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if t, ok := ctx.Value(ctxKeyStopTime).(time.Time); ok {
				elapsed := time.Now().Sub(t).Truncate(time.Millisecond)
				logger.WithContext(ctx).Infof("Stopped %s in %v", appCtx.Name(), elapsed)
			}
			return nil
		},
	})
}
