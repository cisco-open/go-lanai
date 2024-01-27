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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/cliprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/envprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/fileprovider"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"fmt"
	"go.uber.org/fx"
)

type bootstrapProvidersOut struct {
	fx.Out
	ProviderGroup appconfig.ProviderGroup `group:"bootstrap-config"`
}

type appConfigProvidersOut struct {
	fx.Out
	ProviderGroup appconfig.ProviderGroup `group:"application-config"`
}

/*********************
	Bootstrap Groups
 *********************/

func newCommandProviderGroup(execCtx *bootstrap.CliExecContext) bootstrapProvidersOut {
	p := cliprovider.NewCobraProvider(commandlinePrecedence, execCtx, "cli.flag.")
	return bootstrapProvidersOut {
		ProviderGroup: appconfig.NewStaticProviderGroup(commandlinePrecedence, p),
	}
}

func newOsEnvProviderGroup() bootstrapProvidersOut {
	p := envprovider.NewEnvProvider(osEnvPrecedence)
	return bootstrapProvidersOut {
		ProviderGroup: appconfig.NewStaticProviderGroup(osEnvPrecedence, p),
	}
}

func newBootstrapFileProviderGroup() bootstrapProvidersOut {
	const name = "bootstrap"
	const ext = "yml"
	group := appconfig.NewProfileBasedProviderGroup(bootstrapLocalFilePrecedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return name
		}
		return fmt.Sprintf("%s-%s", name, profile)
	}
	group.CreateFunc = func(name string, order int, conf bootstrap.ApplicationConfig) appconfig.Provider {
		ptr, exists := fileprovider.NewFileProvidersFromBaseName(order, name, ext, conf)
		if !exists || ptr == nil {
			return nil
		}
		return ptr
	}
	group.ProcessFunc = func(ctx context.Context, providers []appconfig.Provider) []appconfig.Provider {
		if len(providers) != 0 {
			logger.WithContext(ctx).Infof("Found %d bootstrap configuration files", len(providers))
		}
		return providers
	}

	return bootstrapProvidersOut {
		ProviderGroup: group,
	}
}

type defaultProviderGroupDI struct {
	fx.In
	ExecCtx   *bootstrap.CliExecContext
	Providers []appconfig.Provider `group:"default-config"`
}

func newDefaultProviderGroup(di defaultProviderGroupDI) bootstrapProvidersOut {
	p := cliprovider.NewStaticConfigProvider(defaultPrecedence, di.ExecCtx)
	providers := []appconfig.Provider{p}
	for _, p := range di.Providers {
		if p == nil {
			continue
		}
		if reorder, ok := p.(appconfig.ProviderReorderer); ok {
			reorder.Reorder(defaultPrecedence)
		}
		providers =  append(providers, p)
	}
	return bootstrapProvidersOut {
		ProviderGroup: appconfig.NewStaticProviderGroup(defaultPrecedence, providers...),
	}
}
