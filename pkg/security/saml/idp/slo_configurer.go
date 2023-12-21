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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"net/http"
)

type SamlLogoutEndpointConfigurer struct {
	samlConfigurer
}

func newSamlLogoutEndpointConfigurer(properties samlctx.SamlProperties,
	samlClientStore samlctx.SamlClientStore) *SamlLogoutEndpointConfigurer {

	return &SamlLogoutEndpointConfigurer{
		samlConfigurer: samlConfigurer{
			properties:      properties,
			samlClientStore: samlClientStore,
		},
	}
}

func (c *SamlLogoutEndpointConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	f := feature.(*Feature)
	if len(f.logoutUrl) == 0 {
		// not enabled
		return
	}

	metaMw := c.metadataMiddleware(f)
	mw := NewSamlSingleLogoutMiddleware(metaMw)
	ws.
		Add(middleware.NewBuilder("Saml Service Provider Refresh").
			ApplyTo(matcher.RouteWithPattern(f.logoutUrl, http.MethodGet, http.MethodPost)).
			Order(security.MWOrderSAMLMetadataRefresh).
			Use(mw.RefreshMetadataHandler(mw.SLOCondition())),
		)

	logout.Configure(ws).
		AddLogoutHandler(mw).
		AddSuccessHandler(mw).
		AddErrorHandler(mw).
		AddEntryPoint(mw)

	return nil
}
