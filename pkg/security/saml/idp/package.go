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

package samlidp

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/security"
	samlctx "github.com/cisco-open/go-lanai/pkg/security/saml"
	"github.com/cisco-open/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "saml auth - authorize",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

var logger = log.New("SAML.SSO")

func Use() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	SecRegistrar     security.Registrar `optional:"true"`
	Properties       samlctx.SamlProperties
	ServerProperties web.ServerProperties
	ServiceProviderManager samlctx.SamlClientStore `optional:"true"`
	AccountStore           security.AccountStore `optional:"true"`
	AttributeGenerator     AttributeGenerator `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		authConfigurer := newSamlAuthorizeEndpointConfigurer(di.Properties,
			di.ServiceProviderManager, di.AccountStore,
			di.AttributeGenerator)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, authConfigurer)

		sloConfigurer := newSamlLogoutEndpointConfigurer(di.Properties, di.ServiceProviderManager)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(SloFeatureId, sloConfigurer)
	}
}