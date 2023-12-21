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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"net/http"
)

type SamlAuthorizeEndpointConfigurer struct {
	samlConfigurer
	accountStore       security.AccountStore
	attributeGenerator AttributeGenerator
}

func newSamlAuthorizeEndpointConfigurer(properties samlctx.SamlProperties,
	samlClientStore samlctx.SamlClientStore,
	accountStore security.AccountStore,
	attributeGenerator AttributeGenerator) *SamlAuthorizeEndpointConfigurer {

	return &SamlAuthorizeEndpointConfigurer{
		samlConfigurer: samlConfigurer{
			properties:      properties,
			samlClientStore: samlClientStore,
		},
		accountStore:       accountStore,
		attributeGenerator: attributeGenerator,
	}
}

func (c *SamlAuthorizeEndpointConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	f := feature.(*Feature)

	metaMw := c.metadataMiddleware(f)
	mw := NewSamlAuthorizeEndpointMiddleware(metaMw, c.accountStore, c.attributeGenerator)

	ws.
		Add(middleware.NewBuilder("Saml Service Provider Refresh").
			ApplyTo(matcher.RouteWithPattern(f.ssoLocation.Path, http.MethodGet, http.MethodPost)).
			Order(security.MWOrderSAMLMetadataRefresh).
			Use(mw.RefreshMetadataHandler(f.ssoCondition))).
		Add(middleware.NewBuilder("Saml SSO").
			ApplyTo(matcher.RouteWithPattern(f.ssoLocation.Path, http.MethodGet, http.MethodPost)).
			Order(security.MWOrderSamlAuthEndpoints).
			Use(mw.AuthorizeHandlerFunc(f.ssoCondition)))

	ws.Add(mapping.Get(f.ssoLocation.Path).HandlerFunc(security.NoopHandlerFunc()))
	ws.Add(mapping.Post(f.ssoLocation.Path).HandlerFunc(security.NoopHandlerFunc()))

	//metadata is an actual endpoint
	ws.Add(mapping.Get(f.metadataPath).
		HandlerFunc(mw.MetadataHandlerFunc()).
		Name("saml metadata"))

	// configure error handling
	errorhandling.Configure(ws).
		AdditionalErrorHandler(NewSamlErrorHandler())
	return nil
}
