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
	"embed"
	"github.com/cisco-open/go-lanai/pkg/appconfig"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"go.uber.org/fx"
)

const (
	//preserve gap between different property sources to allow space for profile specific properties.
	precedenceGap = 1000

	//lower integer means higher precedence, therefore the list here is high to low in terms of precedence
	_ = iota * precedenceGap
	PrecedenceExternalAppContext
	PrecedenceExternalDefaultContext
	PrecedenceApplicationAdHoc
	PrecedenceBootstrapAdHoc
	PrecedenceCommandline
	PrecedenceOSEnv
	PrecedenceApplicationLocalFile
	PrecedenceBootstrapLocalFile
	PrecedenceDefault
)

var logger = log.New("Config")

//go:embed defaults-global.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "bootstrap endpoint",
	Precedence: bootstrap.AppConfigPrecedence,
	PriorityOptions: []fx.Option{
		FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(
			// Bootstrap groups and config
			newCommandProviderGroup,
			newOsEnvProviderGroup,
			newBootstrapFileProviderGroup,
			newDefaultProviderGroup,
			newBootstrapAdHocProviderGroup,
			newBootstrapConfig,
			// Application file & adhoc
			newApplicationFileProviderGroup,
			newApplicationAdHocProviderGroup,
			// App Config
			newApplicationConfig,
			newGlobalProperties,
		),
	},
}

// Use Entrypoint of appconfig package
func Use() {
	bootstrap.Register(Module)
}

type bootstrapConfigDI struct {
	fx.In
	App            *bootstrap.App
	ProviderGroups []appconfig.ProviderGroup `group:"bootstrap-config"`
}

func newBootstrapConfig(di bootstrapConfigDI) *appconfig.BootstrapConfig {
	var groups []appconfig.ProviderGroup
	for _, g := range di.ProviderGroups {
		if g != nil {
			groups = append(groups, g)
		}
	}

	bootstrapConfig := appconfig.NewBootstrapConfig(groups...)
	if e := bootstrapConfig.Load(di.App.EagerGetApplicationContext(), false); e != nil {
		panic(e)
	}

	return bootstrapConfig
}

type appConfigDIOut struct {
	fx.Out
	ACPtr *appconfig.ApplicationConfig
	ACI   bootstrap.ApplicationConfig
}

type appConfigDI struct {
	fx.In
	App             *bootstrap.App
	ProviderGroups  []appconfig.ProviderGroup `group:"application-config"`
	BootstrapConfig *appconfig.BootstrapConfig
}

// expose *appconfig.ApplicationConfig as both pointer and interface
func newApplicationConfig(di appConfigDI) appConfigDIOut {
	var groups []appconfig.ProviderGroup
	for _, g := range di.ProviderGroups {
		if g != nil {
			groups = append(groups, g)
		}
	}
	for _, g := range di.BootstrapConfig.ProviderGroups() {
		groups = append(groups, g)
	}

	applicationConfig := appconfig.NewApplicationConfig(groups...)
	if e := applicationConfig.Load(di.App.EagerGetApplicationContext(), false); e != nil {
		panic(e)
	}

	return appConfigDIOut{
		ACPtr: applicationConfig,
		ACI:   applicationConfig,
	}
}

func newGlobalProperties(cfg *appconfig.ApplicationConfig) bootstrap.Properties {
	props := bootstrap.Properties{}
	if e := cfg.Bind(&props, ""); e != nil {
		panic(e)
	}
	return props
}
