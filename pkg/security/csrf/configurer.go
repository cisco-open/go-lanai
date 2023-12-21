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

package csrf

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

/**
	CSRF feature uses the synchronizer token pattern to prevent cross site request forgery
	https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#synchronizer-token-pattern
 */

var FeatureId = security.FeatureId("csrf", security.FeatureOrderCsrf)

type Feature struct {
	requireCsrfProtectionMatchers []web.RequestMatcher
	ignoreCsrfProtectionMatchers  []web.RequestMatcher
	csrfDeniedHandler             security.AccessDeniedHandler
}

func Configure(ws security.WebSecurity) *Feature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return  fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure CSRF: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

func New() *Feature {
	return &Feature{}
}

func (f *Feature) AddCsrfProtectionMatcher(m web.RequestMatcher) *Feature {
	f.requireCsrfProtectionMatchers = append(f.requireCsrfProtectionMatchers, m)
	return f
}

func (f *Feature) IgnoreCsrfProtectionMatcher(m web.RequestMatcher) *Feature {
	f.ignoreCsrfProtectionMatchers = append(f.ignoreCsrfProtectionMatchers, m)
	return f
}

func (f *Feature) CsrfDeniedHandler(csrfDeniedHandler security.AccessDeniedHandler) *Feature {
	f.csrfDeniedHandler = csrfDeniedHandler
	return f
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

type Configurer struct {
}

func newCsrfConfigurer() *Configurer{
	return &Configurer{}
}

func (sc *Configurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	// configure additional access denied handler if provided
	if f.csrfDeniedHandler != nil {
		handler := &CsrfDeniedHandler{delegate: f.csrfDeniedHandler}
		ws.Shared(security.WSSharedKeyCompositeAccessDeniedHandler).(*security.CompositeAccessDeniedHandler).
			Add(handler)
	}

	tokenStore := newSessionBackedStore()

	//Add authentication success handler
	successHandler := &ChangeCsrfHandler{
		csrfTokenStore: tokenStore,
	}
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(successHandler)

	// configure middleware
	requiredCsrfProtectionMatcher := DefaultProtectionMatcher
	for _, matcher := range f.requireCsrfProtectionMatchers {
		requiredCsrfProtectionMatcher = requiredCsrfProtectionMatcher.Or(matcher)
	}

	ignoreCsrfProtectionMatcher := DefaultIgnoreMatcher
	for _, matcher := range f.ignoreCsrfProtectionMatchers {
		ignoreCsrfProtectionMatcher = ignoreCsrfProtectionMatcher.Or(matcher)
	}

	manager := newManager(tokenStore, requiredCsrfProtectionMatcher, ignoreCsrfProtectionMatcher)
	csrfHandler := middleware.NewBuilder("csrfMiddleware").
		Order(security.MWOrderCsrfHandling).
		Use(manager.CsrfHandlerFunc())

	ws.Add(csrfHandler)
	return nil
}

