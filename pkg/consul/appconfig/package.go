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

package consulappconfig

import (
	"github.com/cisco-open/go-lanai/pkg/appconfig"
	appconfiginit "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "bootstrap endpoint",
	Precedence: bootstrap.AppConfigPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(
			// Consul
			bindConsulConfigProperties,
			fxNewConsulDefaultContextProviderGroup,
			fxNewConsulAppContextProviderGroup,
		),
	},
}

type groupDI struct {
	fx.In
	BootstrapConfig        *appconfig.BootstrapConfig
	ConsulConfigProperties ConsulConfigProperties
	ConsulConnection       *consul.Connection `optional:"true"`
}

type appConfigProvidersOut struct {
	fx.Out
	ProviderGroup appconfig.ProviderGroup `group:"application-config"`
}

func withProperties(props *ConsulConfigProperties) ProviderGroupOptions {
	return func(opt *ProviderGroupOption) {
		opt.Prefix = props.Prefix
		opt.Path = props.DefaultContext
		opt.ProfileSeparator = props.ProfileSeparator
	}
}

func fxNewConsulDefaultContextProviderGroup(di groupDI) appConfigProvidersOut {
	if !di.ConsulConfigProperties.Enabled || di.ConsulConnection == nil {
		return appConfigProvidersOut{}
	}

	return appConfigProvidersOut{
		ProviderGroup: NewProviderGroup(withProperties(&di.ConsulConfigProperties),
			func(opt *ProviderGroupOption) {
				opt.Precedence = appconfiginit.PrecedenceExternalDefaultContext
				opt.Connection = di.ConsulConnection
			},
		),
	}
}

func fxNewConsulAppContextProviderGroup(di groupDI) appConfigProvidersOut {
	if !di.ConsulConfigProperties.Enabled || di.ConsulConnection == nil {
		return appConfigProvidersOut{}
	}

	appName, _ := di.BootstrapConfig.Value(bootstrap.PropertyKeyApplicationName).(string)
	return appConfigProvidersOut{
		ProviderGroup: NewProviderGroup(withProperties(&di.ConsulConfigProperties),
			func(opt *ProviderGroupOption) {
				opt.Precedence = appconfiginit.PrecedenceExternalDefaultContext
				opt.Path = appName
				opt.Connection = di.ConsulConnection
			},
		),
	}
}
