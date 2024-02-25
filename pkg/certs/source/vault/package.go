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

package vaultcerts

import (
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/certs"
    certsource "github.com/cisco-open/go-lanai/pkg/certs/source"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/vault"
    "go.uber.org/fx"
)

var logger = log.New("Certs.Vault")

const (
	sourceType = certs.SourceVault
)

var Module = &bootstrap.Module{
	Name:       "certs-vault",
	Precedence: bootstrap.TlsConfigPrecedence,
	Options: []fx.Option{
		fx.Provide(FxProvider()),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type factoryDI struct {
	fx.In
	AppCtx      *bootstrap.ApplicationContext
	Props       certs.Properties `optional:"true"`
	VaultClient *vault.Client    `optional:"true"`
}

func FxProvider() fx.Annotated {
	return fx.Annotated{
		Group: certs.FxGroup,
		Target: func(di factoryDI) (certs.SourceFactory, error) {
			if di.VaultClient == nil {
				logger.Warnf(`Vault Certificates source is not supported. Tips: Do not forget to initialize vault client.`)
				return nil, nil
			}

			var rawDefaults json.RawMessage
			if di.Props.Sources != nil {
				rawDefaults, _ = di.Props.Sources[sourceType]
			}
			factory, e := certsource.NewFactory[SourceProperties](sourceType, rawDefaults, func(props SourceProperties) certs.Source {
				return NewVaultProvider(di.AppCtx, di.VaultClient, props)
			})
			if e != nil {
				return nil, fmt.Errorf(`unable to register certificate source type [%s]: %v`, sourceType, e)
			}
			return factory, nil
		},
	}
}
