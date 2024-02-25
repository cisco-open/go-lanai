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

package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"go.uber.org/fx"
)

var logger = log.New("Security")

var Module = &bootstrap.Module{
	Name: "security",
	Precedence: MaxSecurityPrecedence,
	Options: []fx.Option{
		fx.Provide(provideSecurityInitialization),
		fx.Invoke(initialize),
	},
}

// Use Maker func, does nothing. Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
	template.RegisterGlobalModelValuer(template.ModelKeySecurity, template.ContextModelValuer(Get))
}

/**************************
	Provider
***************************/
type dependencies struct {
	fx.In
	GlobalAuthenticator Authenticator `optional:"true"`
	// may be generic security properties
}

type global struct {
	fx.Out
	Initializer Initializer
	Registrar Registrar
}

// We let configurer.initializer can be autowired as both Initializer and Registrar
func provideSecurityInitialization(di dependencies) global {
	initializer := newSecurity(di.GlobalAuthenticator)
	return global{
		Initializer: initializer,
		Registrar: initializer,
	}
}

/**************************
	Initialize
***************************/
type initDI struct {
	fx.In
	AppContext           *bootstrap.ApplicationContext
	Registerer           *web.Registrar `optional:"true"`
	Initializer          Initializer
}

func initialize(lc fx.Lifecycle, di initDI) {
	if err := di.Initializer.Initialize(di.AppContext, lc, di.Registerer); err != nil {
		panic(err)
	}
}

