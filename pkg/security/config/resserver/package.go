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

package resserver

import (
    "embed"
    appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/timeoutsupport"
    "go.uber.org/fx"
)

//go:embed defaults-resserver.yml
var defaultConfigFS embed.FS

//goland:noinspection GoNameStartsWithPackageName
var Module = &bootstrap.Module{
	Name: "oauth2 authserver",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(jwt.BindCryptoProperties),
		fx.Provide(ProvideResServerDI),
		fx.Invoke(ConfigureResourceServer),
	},
}

func Use() {
	security.Use()
	bootstrap.Register(Module)
	timeoutsupport.Use()
}

