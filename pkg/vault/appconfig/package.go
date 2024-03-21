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

package vaultappconfig

import (
	"github.com/cisco-open/go-lanai/pkg/appconfig"
	appconfiginit "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/vault"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "bootstrap endpoint",
	Precedence: bootstrap.AppConfigPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(
			bindVaultConfigProperties,
			fxNewVaultDefaultContextProviderGroup,
			fxNewVaultAppContextProviderGroup,
		),
	},
}

type groupDI struct {
	fx.In
	BootstrapConfig       *appconfig.BootstrapConfig
	VaultConfigProperties VaultConfigProperties
	VaultClient           *vault.Client `optional:"true"`
}

type appConfigProvidersOut struct {
	fx.Out
	ProviderGroup appconfig.ProviderGroup `group:"application-config"`
}

func withProperties(props *VaultConfigProperties) ProviderGroupOptions {
	return func(opt *ProviderGroupOption) {
		opt.Backend = props.Backend
		opt.BackendVersion = props.BackendVersion
		opt.Path = props.DefaultContext
		opt.ProfileSeparator = props.ProfileSeparator
	}
}

func fxNewVaultDefaultContextProviderGroup(di groupDI) (appConfigProvidersOut, error) {
	if !di.VaultConfigProperties.Enabled || di.VaultClient == nil {
		return appConfigProvidersOut{}, nil
	}

	group, e := NewProviderGroup(withProperties(&di.VaultConfigProperties),
		func(opt *ProviderGroupOption) {
			opt.Precedence = appconfiginit.PrecedenceExternalDefaultContext
			opt.VaultClient = di.VaultClient
		},
	)
	out := appConfigProvidersOut{
		ProviderGroup: group,
	}
	return out, e
}

func fxNewVaultAppContextProviderGroup(di groupDI) (appConfigProvidersOut, error) {
	if !di.VaultConfigProperties.Enabled || di.VaultClient == nil {
		return appConfigProvidersOut{}, nil
	}

	appName, _ := di.BootstrapConfig.Value(bootstrap.PropertyKeyApplicationName).(string)
	group, e := NewProviderGroup(withProperties(&di.VaultConfigProperties),
		func(opt *ProviderGroupOption) {
			opt.Precedence = appconfiginit.PrecedenceExternalAppContext
			opt.Path = appName
			opt.VaultClient = di.VaultClient
		},
	)
	out := appConfigProvidersOut{
		ProviderGroup: group,
	}
	return out, e
}

