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

package sp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"go.uber.org/fx"
)

var logger = log.New("SAML.Auth")

var Module = &bootstrap.Module{
	Name: "saml authenticator",
	Precedence: security.MinSecurityPrecedence + 30,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	gob.Register((*samlAssertionAuthentication)(nil))
}

func Use() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	SecRegistrar   security.Registrar `optional:"true"`
	SamlProperties samlctx.SamlProperties
	ServerProps    web.ServerProperties
	IdpManager     idp.IdentityProviderManager
	AccountStore   security.FederatedAccountStore
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		shared := newSamlConfigurer(di.SamlProperties, di.IdpManager)
		loginConfigurer := newSamlAuthConfigurer(shared, di.AccountStore)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, loginConfigurer)

		logoutConfigurer := newSamlLogoutConfigurer(shared)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(LogoutFeatureId, logoutConfigurer)
	}
}