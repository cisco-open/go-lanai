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

package consulsd

import (
	"context"
	"embed"
	"fmt"
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("Consul.SD")

//go:embed defaults-discovery.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "consul service discovery",
	Precedence: bootstrap.ServiceDiscoveryPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(
			BindDiscoveryProperties,
			fx.Annotate(discovery.NewBuildInfoCustomizer, fxRegistrationCustomizerGroupTag()),
			fx.Annotate(providePropertiesBasedCustomizer, fxRegistrationCustomizerGroupTag()),
			NewServiceRegistrar,
			provideRegistration,
			provideDiscoveryClient),
		fx.Invoke(registerService),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func fxRegistrationCustomizerGroupTag() fx.Annotation {
	return fx.ResultTags(fmt.Sprintf(`group:"%s"`, discovery.FxGroup))
}

func providePropertiesBasedCustomizer(appCtx *bootstrap.ApplicationContext) discovery.ServiceRegistrationCustomizer {
	return discovery.NewPropertiesBasedCustomizer(appCtx, nil)
}

type regDI struct {
	fx.In
	AppCtx      *bootstrap.ApplicationContext
	Props       DiscoveryProperties
	Customizers []discovery.ServiceRegistrationCustomizer `group:"discovery"`
}

func provideRegistration(di regDI) discovery.ServiceRegistration {
	reg := NewRegistration(di.AppCtx,
		RegistrationWithAppContext(di.AppCtx),
		RegistrationWithProperties(&di.Props),
		RegistrationWithCustomizers(di.Customizers...))
	return reg
}

func provideDiscoveryClient(ctx *bootstrap.ApplicationContext, conn *consul.Connection, props DiscoveryProperties) discovery.Client {
	return NewConsulDiscoveryClient(ctx, conn, func(opt *ClientConfig) {
		opt.DefaultSelector = InstanceWithProperties(&props.DefaultSelector)
	})
}

func registerService(lc fx.Lifecycle, registrar discovery.ServiceRegistrar, registration discovery.ServiceRegistration) {
	// because we are the lowest precedence, this is executed when every thing is ready
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return registrar.Register(ctx, registration)
		},
		OnStop: func(ctx context.Context) error {
			return registrar.Deregister(ctx, registration)
		},
	})
}
