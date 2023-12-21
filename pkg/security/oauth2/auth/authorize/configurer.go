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

package authorize

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
)

var (
	FeatureId = security.FeatureId("OAuth2AuthorizeEndpoint", security.FeatureOrderOAuth2AuthorizeEndpoint)
)

//goland:noinspection GoNameStartsWithPackageName
type AuthorizeEndpointConfigurer struct {
}

func newOAuth2AuhtorizeEndpointConfigurer() *AuthorizeEndpointConfigurer {
	return &AuthorizeEndpointConfigurer{
	}
}

func (c *AuthorizeEndpointConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	// Verify
	f := feature.(*AuthorizeFeature)
	if err := c.validate(f, ws); err != nil {
		return err
	}

	// configure other features
	errorhandling.Configure(ws).
		AdditionalErrorHandler(f.errorHandler)

	//prepare middlewares
	authRouteMatcher := matcher.RouteWithPattern(f.path, http.MethodGet, http.MethodPost)
	approveRouteMatcher := matcher.RouteWithPattern(f.approvalPath, http.MethodPost)
	approveRequestMatcher := matcher.RequestWithPattern(f.approvalPath, http.MethodPost).
		And(matcher.RequestHasPostForm(oauth2.ParameterUserApproval))

	authorizeMW := NewAuthorizeEndpointMiddleware(func(opts *AuthorizeMWOption) {
		opts.RequestProcessor = f.requestProcessor
		opts.AuthorizeHandler = f.authorizeHandler
		opts.ApprovalMatcher = approveRequestMatcher
	})

	// install middlewares
	preAuth := middleware.NewBuilder("authorize validation").
		ApplyTo(authRouteMatcher.Or(approveRouteMatcher)).
		Order(security.MWOrderOAuth2AuthValidation).
		Use(authorizeMW.PreAuthenticateHandlerFunc(f.condition))

	ws.Add(preAuth)

	// install authorize endpoint
	epGet := mapping.Get(f.path).Name("authorize GET").
		HandlerFunc(authorizeMW.AuthorizeHandlerFunc(f.condition))
	epPost := mapping.Post(f.path).Name("authorize Post").
		HandlerFunc(authorizeMW.AuthorizeHandlerFunc(f.condition))

	ws.Route(authRouteMatcher).Add(epGet, epPost)

	// install approve endpoint
	approve := mapping.Post(f.approvalPath).Name("approve endpoint").
		HandlerFunc(authorizeMW.ApproveOrDenyHandlerFunc())

	ws.Route(approveRouteMatcher).Add(approve)

	return nil
}

func (c *AuthorizeEndpointConfigurer) validate(f *AuthorizeFeature, ws security.WebSecurity) error {
	if f.path == "" {
		return fmt.Errorf("authorize endpoint path is not set")
	}

	if f.errorHandler == nil {
		f.errorHandler = auth.NewOAuth2ErrorHandler()
	}

	if f.authorizeHandler == nil {
		return fmt.Errorf("auhtorize handler is not set")
	}

	//if f.granters == nil || len(f.granters) == 0 {
	//	return fmt.Errorf("token granters is not set")
	//}
	return nil
}



