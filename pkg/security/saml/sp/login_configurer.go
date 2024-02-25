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
	"github.com/cisco-open/go-lanai/pkg/security/access"
	"github.com/cisco-open/go-lanai/pkg/security/errorhandling"
	"github.com/cisco-open/go-lanai/pkg/security/request_cache"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"github.com/cisco-open/go-lanai/pkg/web/mapping"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
	"github.com/cisco-open/go-lanai/pkg/web/middleware"
)

type SamlAuthConfigurer struct {
	*samlConfigurer
	accountStore   security.FederatedAccountStore
}

func (c *SamlAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	m := c.makeMiddleware(f, ws)

	ws.Route(matcher.RouteWithPattern(f.acsPath)).
		Route(matcher.RouteWithPattern(f.metadataPath)).
		Add(mapping.Get(f.metadataPath).
			HandlerFunc(m.MetadataHandlerFunc()).
			//metadata is an endpoint that is available without conditions, therefore call Build() to not inherit the ws condition
			Name("saml metadata").Build()).
		Add(mapping.Post(f.acsPath).
			HandlerFunc(m.ACSHandlerFunc()).
			Name("saml assertion consumer m")).
		Add(middleware.NewBuilder("saml idp metadata refresh").
			Order(security.MWOrderSAMLMetadataRefresh).
			Use(m.RefreshMetadataHandler()))

	requestMatcher := matcher.RequestWithPattern(f.acsPath).Or(matcher.RequestWithPattern(f.metadataPath))
	access.Configure(ws).
	Request(requestMatcher).WithOrder(order.Highest).PermitAll()

	//authentication entry point
	errorhandling.Configure(ws).
		AuthenticationEntryPoint(request_cache.NewSaveRequestEntryPoint(m))
	return nil
}

func (c *SamlAuthConfigurer) makeMiddleware(f *Feature, ws security.WebSecurity) *SPLoginMiddleware {
	opts := c.getServiceProviderConfiguration(f)
	sp := c.sharedServiceProvider(opts)
	clientManager := c.sharedClientManager(opts)
	tracker := c.sharedRequestTracker(opts)
	if f.successHandler == nil {
		f.successHandler = NewTrackedRequestSuccessHandler(tracker)
	}

	authenticator := &Authenticator{
		accountStore: c.accountStore,
		idpManager:   c.samlIdpManager,
	}


	return NewLoginMiddleware(sp, tracker, c.idpManager, clientManager, c.effectiveSuccessHandler(f, ws), authenticator, f.errorPath)
}

func newSamlAuthConfigurer(shared *samlConfigurer, accountStore security.FederatedAccountStore) *SamlAuthConfigurer {
	return &SamlAuthConfigurer{
		samlConfigurer: shared,
		accountStore:   accountStore,
	}
}