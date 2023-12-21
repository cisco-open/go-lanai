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

package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/extsamlidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/unknownIdp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/assets"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"embed"
	"go.uber.org/fx"
	"net/url"
)

var logger = log.New("SEC.Example")

//go:generate npm install --prefix web/nodejs
//go:generate go run github.com/mholt/archiver/v3/cmd/arc -overwrite -folder-safe=false unarchive web/nodejs/node_modules/@msx/login-app/login-app-ui.zip web/login-ui/
//go:embed web/login-ui/*
var GeneratedContent embed.FS

// Maker func, does nothing. Allow service to include this module in main()
func Use() {
	authserver.Use()
	resserver.Use()
	bootstrap.AddOptions(
		fx.Provide(BindAccountsProperties),
		fx.Provide(BindAccountPoliciesProperties),
		fx.Provide(BindClientsProperties),
		fx.Provide(NewInMemoryAccountStore),
		fx.Provide(NewInMemoryFederatedAccountStore),
		fx.Provide(NewInMemoryClientStore),
		fx.Provide(NewTenantStore),
		fx.Provide(NewProviderStore),
		fx.Provide(NewInMemoryIdpManager),
		fx.Provide(NewInMemSpManager),
		fx.Provide(newAuthServerConfigurer),
		fx.Invoke(configureWeb),
		fx.Invoke(configureSecurity),
		fx.Invoke(configureConsulRegistration),
	)
}

func configureSecurity(init security.Registrar) {
	init.Register(&ErrorPageSecurityConfigurer{})
}

type authDI struct {
	fx.In
	ClientStore   oauth2.OAuth2ClientStore
	AccountStore  security.AccountStore
	TenantStore   security.TenantStore
	ProviderStore security.ProviderStore
	IdpManager    idp.IdentityProviderManager
}

func newAuthServerConfigurer(di authDI) authserver.AuthorizationServerConfigurer {
	return func(config *authserver.Configuration) {
		config.AddIdp(passwdidp.NewPasswordIdpSecurityConfigurer())
		config.AddIdp(extsamlidp.NewSamlIdpSecurityConfigurer())
		config.AddIdp(unknownIdp.NewNoIdpSecurityConfigurer())

		config.IdpManager = di.IdpManager
		config.ClientStore = di.ClientStore
		config.ClientSecretEncoder = passwd.NewNoopPasswordEncoder()
		config.UserAccountStore = di.AccountStore
		config.TenantStore = di.TenantStore
		config.ProviderStore = di.ProviderStore
		config.UserPasswordEncoder = passwd.NewNoopPasswordEncoder()
		config.Endpoints = authserver.Endpoints{
			Authorize: authserver.ConditionalEndpoint{
				Location: &url.URL{Path: "/v2/authorize"},
				Condition: matcher.NotRequest(matcher.RequestWithForm("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer")),
			},
			Approval: "/v2/approve",
			Token: "/v2/token",
			CheckToken: "/v2/check_token",
			UserInfo: "/v2/userinfo",
			JwkSet: "/v2/jwks",
			Logout: "/v2/logout",
			SamlSso: authserver.ConditionalEndpoint{
				Location: &url.URL{Path:"/v2/authorize", RawQuery: "grant_type=urn:ietf:params:oauth:grant-type:saml2-bearer"},
				Condition: matcher.RequestWithForm("grant_type", "urn:ietf:params:oauth:grant-type:saml2-bearer"),
			},
			SamlMetadata: "/metadata",
			TenantHierarchy: "/v2/tenant_hierarchy",
		}
	}
}

func configureWeb(r *web.Registrar) {
	r.MustRegister(web.OrderedFS(GeneratedContent, passwdidp.OrderTemplateFSOverwrite))
	r.MustRegister(assets.New("app", "web/login-ui"))
	r.MustRegister(NewLoginFormController())
}

func configureConsulRegistration(r *discovery.Customizers) {
	r.Add(&RegistrationCustomizer{})
}