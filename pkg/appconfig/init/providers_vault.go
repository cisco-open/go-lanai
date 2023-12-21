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

package appconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/vaultprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"fmt"
	"go.uber.org/fx"
)

type vaultDi struct {
	fx.In
	BootstrapConfig       *appconfig.BootstrapConfig
	VaultConfigProperties *vaultprovider.KvConfigProperties
	VaultClient           *vault.Client `optional:"true"`
}

func newVaultDefaultContextProviderGroup(di vaultDi) appConfigProvidersOut {
	if !di.VaultConfigProperties.Enabled || di.VaultClient == nil{
		return appConfigProvidersOut{}
	}

	kvSecretEngine, err := vaultprovider.NewKvSecretEngine(
		di.VaultConfigProperties.BackendVersion, di.VaultConfigProperties.Backend, di.VaultClient)

	if err != nil {
		panic(err)
	}

	group := appconfig.NewProfileBasedProviderGroup(externalDefaultContextPrecedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return di.VaultConfigProperties.DefaultContext
		}
		return fmt.Sprintf("%s%s%s", di.VaultConfigProperties.DefaultContext, di.VaultConfigProperties.ProfileSeparator, profile)
	}
	group.CreateFunc = func(name string, order int, _ bootstrap.ApplicationConfig) appconfig.Provider {
		return vaultprovider.NewVaultKvProvider(order, name, kvSecretEngine)
	}

	return appConfigProvidersOut {
		ProviderGroup: group,
	}
}

func newVaultAppContextProviderGroup(di vaultDi) appConfigProvidersOut {
	if !di.VaultConfigProperties.Enabled || di.VaultClient == nil{
		return appConfigProvidersOut{}
	}

	kvSecretEngine, err := vaultprovider.NewKvSecretEngine(
		di.VaultConfigProperties.BackendVersion, di.VaultConfigProperties.Backend, di.VaultClient)

	if err != nil {
		panic(err)
	}

	appName := di.BootstrapConfig.Value(bootstrap.PropertyKeyApplicationName)

	group := appconfig.NewProfileBasedProviderGroup(externalAppContextPrecedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return fmt.Sprintf("%s", appName)
		}
		return fmt.Sprintf("%s%s%s", appName, di.VaultConfigProperties.ProfileSeparator, profile)
	}
	group.CreateFunc = func(name string, order int, _ bootstrap.ApplicationConfig) appconfig.Provider {
		ptr := vaultprovider.NewVaultKvProvider(order, name, kvSecretEngine)
		if ptr == nil {
			return nil
		}
		return ptr
	}

	return appConfigProvidersOut {
		ProviderGroup: group,
	}
}
