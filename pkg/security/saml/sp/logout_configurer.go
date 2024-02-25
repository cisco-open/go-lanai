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
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/csrf"
	"github.com/cisco-open/go-lanai/pkg/security/logout"
	"github.com/cisco-open/go-lanai/pkg/security/redirect"
	"github.com/cisco-open/go-lanai/pkg/security/request_cache"
	"github.com/cisco-open/go-lanai/pkg/web/mapping"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
	"github.com/cisco-open/go-lanai/pkg/web/middleware"
)

type SamlLogoutConfigurer struct {
	*samlConfigurer
}

func (c *SamlLogoutConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	m := c.makeMiddleware(f, ws)
	lh := c.makeLogoutHandler(f, ws)
	ep := request_cache.NewSaveRequestEntryPoint(m)

	// configure on top of existing logout feature
	logout.Configure(ws).
		AddLogoutHandler(lh).
		AddEntryPoint(ep)

	// Add some additional endpoints.
	// Note: those endpoints are available regardless what auth method is used, so no condition is applied
	ws.Route(matcher.RouteWithPattern(f.sloPath)).
		Add(mapping.Get(f.sloPath).
			HandlerFunc(m.LogoutHandlerFunc()).
			Name("saml slo as sp - get"),
		).
		Add(mapping.Post(f.sloPath).
			HandlerFunc(m.LogoutHandlerFunc()).
			Name("saml slo as sp - post"),
		).
		Add(middleware.NewBuilder("saml idp metadata refresh").
			Order(security.MWOrderSAMLMetadataRefresh).
			Use(m.RefreshMetadataHandler()),
		)

	csrf.Configure(ws).
		IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(f.sloPath))
	return nil
}

func (c *SamlLogoutConfigurer) makeLogoutHandler(_ *Feature, _ security.WebSecurity) *SingleLogoutHandler {
	return NewSingleLogoutHandler()
}

func (c *SamlLogoutConfigurer) makeMiddleware(f *Feature, ws security.WebSecurity) *SPLogoutMiddleware {
	opts := c.getServiceProviderConfiguration(f)
	sp := c.sharedServiceProvider(opts)
	clientManager := c.sharedClientManager(opts)
	if f.successHandler == nil {
		f.successHandler = request_cache.NewSavedRequestAuthenticationSuccessHandler(
			redirect.NewRedirectWithURL("/"),
			func(from, to security.Authentication) bool {
				return true
			},
		)
	}

	return NewLogoutMiddleware(sp, c.idpManager, clientManager, c.effectiveSuccessHandler(f, ws))
}

func newSamlLogoutConfigurer(shared *samlConfigurer) *SamlLogoutConfigurer {
	return &SamlLogoutConfigurer{
		samlConfigurer: shared,
	}
}
