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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/appconfig"
    "github.com/cisco-open/go-lanai/pkg/appconfig/fileprovider"
    "go.uber.org/fx"
    "path/filepath"
    "reflect"
)

const (
	FxGroupBootstrap = "bootstrap-config"
	FxGroupApplication = "application-config"
	FxGroupDefaults = "default-config"
)



// FxEmbeddedDefaults returns a specialized fx.Option that take a given embed.FS and load *.yml as default properties
func FxEmbeddedDefaults(fs embed.FS) fx.Option {
	return embeddedFileProviderFxOptions(FxGroupDefaults, fs)
}

// FxEmbeddedApplicationAdHoc returns a specialized fx.Option that take a given embed.FS and load *.yml as application properties
func FxEmbeddedApplicationAdHoc(fs embed.FS) fx.Option {
	return embeddedFileProviderFxOptions(FxGroupApplication, fs)
}

// FxEmbeddedBootstrapAdHoc returns a specialized fx.Option that take a given embed.FS and load *.yml as bootstrap properties
func FxEmbeddedBootstrapAdHoc(fs embed.FS) fx.Option {
	return embeddedFileProviderFxOptions(FxGroupBootstrap, fs)
}

// FxProvideDefaults wraps given interface{} as a fx.Provide of appconfig.Provider with order of default properties
// Supported interface are
// 	- appconfig.Provider
//  - a function that returns/create appconfig.Provider
func FxProvideDefaults(providers ...interface{}) fx.Option {
	return providerFxOptions(FxGroupDefaults, providers)
}

// FxProvideApplicationAdHoc wraps given interface{} as a fx.Provide of appconfig.Provider with order of overriding application properties
// Supported interface are
// 	- appconfig.Provider
//  - a function that returns/create appconfig.Provider
func FxProvideApplicationAdHoc(providers ...interface{}) fx.Option {
	return providerFxOptions(FxGroupApplication, providers)
}

// FxProvideBootstrapAdHoc wraps given interface{} as a fx.Provide of appconfig.Provider with order of overriding bootstrap properties
// Supported interface are
// 	- appconfig.Provider
//  - a function that returns/create appconfig.Provider
func FxProvideBootstrapAdHoc(providers ...interface{}) fx.Option {
	return providerFxOptions(FxGroupBootstrap, providers)
}

func providerFxOptions(fxGroup string, providers []interface{}) fx.Option {
	annotated := make([]interface{}, len(providers))
	for i, p := range providers {
		var target interface{}
		switch provider := p.(type) {
		case appconfig.Provider:
			target = func() appconfig.Provider{
				return provider
			}
		default:
			v := reflect.ValueOf(p)
			if v.Kind() != reflect.Func {
				e := fmt.Errorf("invalid appconfig.FxProvide...() parameters. Support appconfig.Provider or a provide function, but got %T", p)
				panic(e)
			}
			target = p
		}

		annotated[i] = fx.Annotated {
			Group: fxGroup,
			Target: target,
		}
	}
	return fx.Provide(annotated...)
}

func embeddedFileProviderFxOptions(fxGroup string, fs embed.FS) fx.Option {
	files, e := fs.ReadDir(".")
	if e != nil {
		return fx.Supply()
	}

	const ext = "yml"
	providers := make([]interface{}, 0)
	for _, f := range files {
		if !f.IsDir() || filepath.Ext(f.Name()) == ext {
			providers = append(providers, fxEmbeddedFileProvider(fxGroup, f.Name(), fs))
		}
	}
	return fx.Provide(providers...)
}

func fxEmbeddedFileProvider(fxGroup string, filepath string, fs embed.FS) fx.Annotated {
	fn := func() appconfig.Provider{
		// Note order will be overwritten by corresponding provider group
		provider, _ := fileprovider.NewEmbeddedFSProvider(0, filepath, fs)
		return provider
	}

	return fx.Annotated {
		Group: fxGroup,
		Target: fn,
	}
}
