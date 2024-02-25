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
	"github.com/cisco-open/go-lanai/pkg/appconfig"
	"go.uber.org/fx"
)

type adhocBootstrapDI struct {
	fx.In
	Providers []appconfig.Provider `group:"bootstrap-config"`
}

func newBootstrapAdHocProviderGroup(di adhocBootstrapDI) bootstrapProvidersOut {
	providers := make([]appconfig.Provider, 0)
	for _, p := range di.Providers {
		if p == nil {
			continue
		}
		if reorder, ok := p.(appconfig.ProviderReorderer); ok {
			reorder.Reorder(bootstrapAdHocPrecedence)
		}
		providers =  append(providers, p)
	}
	return bootstrapProvidersOut {
		ProviderGroup: appconfig.NewStaticProviderGroup(bootstrapAdHocPrecedence, providers...),
	}
}

type adhocApplicationDI struct {
	fx.In
	Providers []appconfig.Provider `group:"application-config"`
}

func newApplicationAdHocProviderGroup(di adhocApplicationDI) appConfigProvidersOut {
	providers := make([]appconfig.Provider, 0)
	for _, p := range di.Providers {
		if p == nil {
			continue
		}
		if reorder, ok := p.(appconfig.ProviderReorderer); ok {
			reorder.Reorder(applicationAdHocPrecedence)
		}
		providers =  append(providers, p)
	}
	return appConfigProvidersOut {
		ProviderGroup: appconfig.NewStaticProviderGroup(applicationAdHocPrecedence, providers...),
	}
}
