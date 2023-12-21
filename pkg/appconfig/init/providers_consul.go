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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/consulprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"fmt"
	"go.uber.org/fx"
)

type consulDi struct {
	fx.In
	BootstrapConfig *appconfig.BootstrapConfig
	ConsulConfigProperties *consulprovider.ConsulConfigProperties
	ConsulConnection *consul.Connection `optional:"true"`
}

func newConsulDefaultContextProviderGroup(di consulDi) appConfigProvidersOut {
	if !di.ConsulConfigProperties.Enabled || di.ConsulConnection == nil {
		return appConfigProvidersOut{}
	}

	group := appconfig.NewProfileBasedProviderGroup(externalDefaultContextPrecedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return fmt.Sprintf("%s/%s", di.ConsulConfigProperties.Prefix, di.ConsulConfigProperties.DefaultContext)
		}
		return fmt.Sprintf("%s/%s%s%s",
			di.ConsulConfigProperties.Prefix, di.ConsulConfigProperties.DefaultContext, di.ConsulConfigProperties.ProfileSeparator, profile)
	}

	group.CreateFunc = func(name string, order int, _ bootstrap.ApplicationConfig) appconfig.Provider {
		ptr := consulprovider.NewConsulProvider(order, name, di.ConsulConnection)
		if ptr == nil {
			return nil
		}
		return ptr
	}

	return appConfigProvidersOut {
		ProviderGroup: group,
	}
}

func newConsulAppContextProviderGroup(di consulDi) appConfigProvidersOut {
	if !di.ConsulConfigProperties.Enabled || di.ConsulConnection == nil {
		return appConfigProvidersOut{}
	}

	appName := di.BootstrapConfig.Value(bootstrap.PropertyKeyApplicationName)

	group := appconfig.NewProfileBasedProviderGroup(externalAppContextPrecedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return fmt.Sprintf("%s/%s", di.ConsulConfigProperties.Prefix, appName)
		}
		return fmt.Sprintf("%s/%s%s%s",
			di.ConsulConfigProperties.Prefix, appName, di.ConsulConfigProperties.ProfileSeparator, profile)
	}
	group.CreateFunc = func(name string, order int, _ bootstrap.ApplicationConfig) appconfig.Provider {
		ptr := consulprovider.NewConsulProvider(order, name, di.ConsulConnection)
		if ptr == nil {
			return nil
		}
		return ptr
	}

	return appConfigProvidersOut {
		ProviderGroup: group,
	}
}
