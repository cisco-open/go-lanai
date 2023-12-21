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

package authserver

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/timeoutsupport"
	samlidp "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/idp"
	th_loader "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy/loader"
	"embed"
	"go.uber.org/fx"
)

//go:embed defaults-authserver.yml
var defaultConfigFS embed.FS

//goland:noinspection GoNameStartsWithPackageName
var Module = &bootstrap.Module{
	Name:       "oauth2 authserver",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(BindAuthServerProperties),
		fx.Provide(ProvideAuthServerDI),
		fx.Provide(provide),
		fx.Invoke(ConfigureAuthorizationServer),
	},
}

func Use() {
	security.Use()
	th_loader.Use()
	samlidp.Use() // saml_auth enables SAML SSO/SLO
	bootstrap.Register(Module)
	timeoutsupport.Use()
	// Note: External SAML IDP support (samllogin package) is enabled as part of samlidp
}
