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

package discovery

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"embed"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
)

//go:embed defaults-discovery.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module {
	Name: "service discovery",
	Precedence: bootstrap.ServiceDiscoveryPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(discovery.BindDiscoveryProperties,
			discovery.NewCustomizers,
			provideRegistration,
			provideDiscoveryClient),
		fx.Invoke(setupServiceRegistration),
	},
}

func init() {
	bootstrap.Register(Module)
}

func Use() {
	// does nothing. Allow service to include this module in main()
}

type regDI struct {
	fx.In
	AppContext          *bootstrap.ApplicationContext
	DiscoveryProperties discovery.DiscoveryProperties
}

func provideRegistration(di regDI) *api.AgentServiceRegistration {
	return discovery.NewRegistration(discovery.RegistrationWithProperties(di.AppContext, di.DiscoveryProperties))
}

func provideDiscoveryClient(ctx *bootstrap.ApplicationContext, conn *consul.Connection, props discovery.DiscoveryProperties) discovery.Client {
	return discovery.NewConsulDiscoveryClient(ctx, conn, func(opt *discovery.ClientConfig) {
		opt.DefaultSelector = discovery.InstanceWithProperties(&props.DefaultSelector)
	})
}

func setupServiceRegistration(lc fx.Lifecycle,
	connection *consul.Connection, registration *api.AgentServiceRegistration, customizers *discovery.Customizers) {

	//because we are the lowest precendence, we execute when every thing is ready
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			customizers.Apply(ctx, registration)
			_ = discovery.Register(ctx, connection, registration)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			customizers.Apply(ctx, registration)
			_ = discovery.Deregister(ctx, connection, registration)
			return nil
		},
	})
}
