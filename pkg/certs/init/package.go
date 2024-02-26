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

// Package certsinit
// Initialize certificate manager with various of certificate sources
package certsinit

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/certs"
    filecerts "github.com/cisco-open/go-lanai/pkg/certs/source/file"
    "go.uber.org/fx"
    "io"
)

const PropertiesPrefix = `certificates`

var Module = &bootstrap.Module{
	Name:       "certs",
	Precedence: bootstrap.TlsConfigPrecedence,
	Options: []fx.Option{
		fx.Provide(BindProperties, ProvideDefaultManager),
		fx.Provide(
			filecerts.FxProvider(),
		),
		fx.Invoke(RegisterManagerLifecycle),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type mgrDI struct {
	fx.In
	AppCfg    bootstrap.ApplicationConfig
	Props     certs.Properties
	Factories []certs.SourceFactory `group:"certs"`
}

func ProvideDefaultManager(di mgrDI) (certs.Manager, certs.Registrar) {
	reg := certs.NewDefaultManager(func(mgr *certs.DefaultManager) {
		mgr.ConfigLoaderFunc = di.AppCfg.Bind
		mgr.Properties = di.Props
	})
	for _, f := range di.Factories {
		if f != nil {
			reg.MustRegister(f)
		}
	}
	return reg, reg
}

// BindProperties create and bind SessionProperties, with a optional prefix
func BindProperties(appCfg bootstrap.ApplicationConfig) certs.Properties {
	props := certs.NewProperties()
	if e := appCfg.Bind(props, PropertiesPrefix); e != nil {
		panic(fmt.Errorf("failed to bind certificate properties: %v", e))
	}
	return *props
}

func RegisterManagerLifecycle(lc fx.Lifecycle, m certs.Manager) {
	lc.Append(fx.StopHook(func(context.Context) error {
		if closer, ok := m.(io.Closer); ok {
			return closer.Close()
		}
		return nil
	}))
}
