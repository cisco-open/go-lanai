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

package webtest

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var mockedWebModule = &bootstrap.Module{
	Name: "web",
	Precedence: web.MinWebPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(
			web.BindServerProperties,
			web.NewEngine,
			web.NewRegistrar),
		fx.Invoke(initialize),
	},
}


type initDI struct {
	fx.In
	Registrar        *web.Registrar
	Properties       web.ServerProperties
	Controllers      []web.Controller      `group:"controllers"`
	Customizers      []web.Customizer      `group:"customizers"`
	ErrorTranslators []web.ErrorTranslator `group:"error_translators"`
}

func initialize(lc fx.Lifecycle, di initDI) {
	di.Registrar.MustRegister(web.NewLoggingCustomizer(di.Properties))
	di.Registrar.MustRegister(web.NewRecoveryCustomizer())
	di.Registrar.MustRegister(web.NewGinErrorHandlingCustomizer())

	di.Registrar.MustRegister(di.Controllers)
	di.Registrar.MustRegister(di.Customizers)
	di.Registrar.MustRegister(di.ErrorTranslators)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			if err = di.Registrar.Initialize(ctx); err != nil {
				return
			}
			defer func(ctx context.Context) {
				_ = di.Registrar.Cleanup(ctx)
			}(ctx)
			return nil
		},
	})
}
