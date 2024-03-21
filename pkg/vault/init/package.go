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

package vault

import (
    "context"
    "embed"
    "github.com/cisco-open/go-lanai/pkg/appconfig"
    appconfigInit "github.com/cisco-open/go-lanai/pkg/appconfig/init"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/vault"
	vaultappconfig "github.com/cisco-open/go-lanai/pkg/vault/appconfig"
	vaulthealth "github.com/cisco-open/go-lanai/pkg/vault/health"
	vaulttracing "github.com/cisco-open/go-lanai/pkg/vault/tracing"
	"go.uber.org/fx"
)

//go:embed defaults-vault.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "vault",
	Precedence: bootstrap.VaultPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(BindConnectionProperties, ProvideDefaultClient),
	},
	Options: []fx.Option{
		appconfigInit.FxEmbeddedDefaults(defaultConfigFS),
		fx.Invoke(vaulthealth.Register, manageClientLifecycle),
	},
	Modules: []*bootstrap.Module{
		vaultappconfig.Module,
		vaulttracing.Module,
	},
}

// Use func, does nothing. Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

func BindConnectionProperties(bootstrapConfig *appconfig.BootstrapConfig) (vault.ConnectionProperties, error) {
	c := vault.ConnectionProperties{
		Host:           "localhost",
		Port:           8200,
		Scheme:         "http",
		Authentication: vault.Token,
		Token:          "replace_with_token_value",
	}
	if e := bootstrapConfig.Bind(&c, vault.PropertyPrefix); e != nil {
		return c, e
	}
	return c, nil
}

type clientDI struct {
	fx.In
	Props       vault.ConnectionProperties
	Customizers []vault.Options `group:"vault"`
}

func ProvideDefaultClient(di clientDI) *vault.Client {
	opts := append([]vault.Options{
		vault.WithProperties(di.Props),
	}, di.Customizers...)
	client, err := vault.New(opts...)
	if err != nil {
		panic(err)
	}
	return client
}

type lcDI struct {
	fx.In
	AppCtx      *bootstrap.ApplicationContext
	Lifecycle   fx.Lifecycle
	VaultClient *vault.Client `optional:"true"`
}

func manageClientLifecycle(di lcDI) {
	if di.VaultClient == nil {
		return
	}
	di.Lifecycle.Append(fx.StartHook(func(_ context.Context) {
		//nolint:contextcheck // Non-inherited new context - intentional. Start hook context expires when startup finishes
		di.VaultClient.AutoRenewToken(di.AppCtx)
	}))
	di.Lifecycle.Append(fx.StopHook(func(_ context.Context) error {
		return di.VaultClient.Close()
	}))
}

