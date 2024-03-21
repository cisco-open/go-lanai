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

package init

import (
	"context"
	"embed"
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/cors"
	webtracing "github.com/cisco-open/go-lanai/pkg/web/tracing"
	"go.uber.org/fx"
)

//go:embed defaults-web.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "web",
	Precedence: web.MinWebPrecedence,
	PriorityOptions: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(
			web.BindServerProperties,
			web.NewEngine,
			web.NewRegistrar),
		fx.Invoke(setup),
	},
	Modules: []*bootstrap.Module{
		cors.Module, webtracing.Module,
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

/**************************
	Provide dependencies
***************************/

/*
*************************

	Setup

**************************
*/
type initDI struct {
	fx.In
	Registrar        *web.Registrar
	Properties       web.ServerProperties
	Controllers      []web.Controller      `group:"controllers"`
	Customizers      []web.Customizer      `group:"customizers"`
	ErrorTranslators []web.ErrorTranslator `group:"error_translators"`
}

func setup(lc fx.Lifecycle, di initDI) {
	di.Registrar.MustRegister(web.NewLoggingCustomizer(di.Properties))
	di.Registrar.MustRegister(web.NewRecoveryCustomizer())
	di.Registrar.MustRegister(web.NewGinErrorHandlingCustomizer())

	di.Registrar.MustRegister(di.Controllers)
	di.Registrar.MustRegister(di.Customizers)
	di.Registrar.MustRegister(di.ErrorTranslators)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			return di.Registrar.Run(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return di.Registrar.Stop(ctx)
		},
	})
}
