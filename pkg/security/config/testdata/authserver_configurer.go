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

package testdata

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/config/authserver"
	"github.com/cisco-open/go-lanai/pkg/security/idp"
	"github.com/cisco-open/go-lanai/pkg/security/idp/extsamlidp"
	"github.com/cisco-open/go-lanai/pkg/security/idp/passwdidp"
	"github.com/cisco-open/go-lanai/pkg/security/idp/unknownIdp"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
	"github.com/cisco-open/go-lanai/pkg/security/passwd"
	"github.com/cisco-open/go-lanai/test/samltest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"go.uber.org/fx"
)

const (
	IdpDomainPasswd    = "passwd.lanai.com"
	IdpDomainExtSAML   = "saml.lanai.com"
	ExtSamlIdpName     = "ext-saml-idp"
	ExtSamlIdpEntityID = "http://external.saml.com/samlidp/metadata"
	ExtSamlIdpSSOUrl   = "http://external.saml.com/samlidp/authorize"
	ExtSamlIdpSLOUrl   = "http://external.saml.com/samlidp/logout"
)

type authDI struct {
	fx.In
	MockingProperties   sectest.MockingProperties
	IdpManager          idp.IdentityProviderManager
	AccountStore        security.AccountStore
	PasswordEncoder     passwd.PasswordEncoder
	Properties          authserver.AuthServerProperties
	PasswdIDPProperties passwdidp.PwdAuthProperties
	SamlIDPProperties   extsamlidp.SamlAuthProperties
	CustomTokenGranter  auth.TokenGranter          `optional:"true"`
	CustomTokenEnhancer auth.TokenEnhancer         `optional:"true"`
	CustomAuthRegistry  auth.AuthorizationRegistry `optional:"true"`
}

func NewAuthServerConfigurer(di authDI) authserver.AuthorizationServerConfigurer {
	return func(config *authserver.Configuration) {
		// setup IDPs
		config.AddIdp(passwdidp.NewPasswordIdpSecurityConfigurer(
			passwdidp.WithProperties(&di.PasswdIDPProperties),
			passwdidp.WithMFAListeners(),
		))
		config.AddIdp(extsamlidp.NewSamlIdpSecurityConfigurer(
			extsamlidp.WithProperties(&di.SamlIDPProperties),
		))
		config.AddIdp(unknownIdp.NewNoIdpSecurityConfigurer())

		config.IdpManager = di.IdpManager
		config.ClientStore = sectest.NewMockedClientStore(di.MockingProperties.Clients.Values()...)
		config.ClientSecretEncoder = di.PasswordEncoder
		config.UserAccountStore = di.AccountStore
		config.TenantStore = sectest.NewMockedTenantStore(di.MockingProperties.Tenants.Values()...)
		config.ProviderStore = sectest.MockedProviderStore{}
		config.UserPasswordEncoder = di.PasswordEncoder
		config.SessionSettingService = StaticSessionSettingService(1)
		if di.CustomTokenEnhancer != nil {
			config.CustomTokenEnhancer = []auth.TokenEnhancer{di.CustomTokenEnhancer}
		}
		if di.CustomTokenGranter != nil {
			config.CustomTokenGranter = []auth.TokenGranter{di.CustomTokenGranter}
		}
		if di.CustomAuthRegistry != nil {
			config.CustomAuthRegistry = di.CustomAuthRegistry
		}
	}
}

type StaticSessionSettingService int

func (s StaticSessionSettingService) GetMaximumSessions(ctx context.Context) int {
	return int(s)
}

func NewMockedIDPManager() *samltest.MockedIdpManager {
	return samltest.NewMockedIdpManager(func(opt *samltest.IdpManagerMockOption) {
		opt.IDPList = []idp.IdentityProvider{
			extsamlidp.NewIdentityProvider(func(opt *extsamlidp.SamlIdpDetails) {
				opt.EntityId = ExtSamlIdpEntityID
				opt.Domain = IdpDomainExtSAML
				opt.ExternalIdpName = ExtSamlIdpName
				opt.ExternalIdName = "username"
				opt.MetadataLocation = "testdata/ext-saml-metadata.xml"
			}),
		}
		opt.Delegates = []idp.IdentityProviderManager{
			sectest.NewMockedIDPManager(func(opt *sectest.IdpManagerMockOption) {
				opt.PasswdIDPDomain = IdpDomainPasswd
			}),
		}
	})
}
