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

package pqcrypt

import (
    "embed"
    "fmt"
    appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/vault"
    "go.uber.org/fx"
)

//var logger = log.New("Data.Enc")

//go:embed defaults-data-enc.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name:       "data-encryption",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(BindDataEncryptionProperties, provideEncryptor),
		fx.Invoke(initialize),
	},
}

func Use() {
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

type encDI struct {
	fx.In
	Properties DataEncryptionProperties `optional:"true"`
	Client     *vault.Client            `optional:"true"`
	UnnamedEnc Encryptor                `optional:"true"`
}

type encOut struct {
	fx.Out
	Enc Encryptor `name:"data/Encryptor"`
}

func provideEncryptor(di encDI) encOut {
	if di.UnnamedEnc != nil {
		return encOut{
			Enc: di.UnnamedEnc,
		}
	}

	var enc Encryptor
	switch {
	case di.Properties.Enabled:
		if di.Client == nil {
			panic(fmt.Errorf("data encryption enabled but vault client is not initialized"))
		}
		venc := newVaultEncryptor(di.Client, &di.Properties.Key)
		enc = compositeEncryptor{venc, plainTextEncryptor{}}
	default:
		enc = plainTextEncryptor{}
	}
	return encOut{
		Enc: enc,
	}
}

/**************************
	Initialize
***************************/
type initDI struct {
	fx.In
	Enc Encryptor `name:"data/Encryptor"`
}

func initialize(di initDI) {
	encryptor = di.Enc
}
