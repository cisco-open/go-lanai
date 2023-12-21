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

package token

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
)

var (
	FeatureId = security.FeatureId("OAuth2AuthToken", security.FeatureOrderOAuth2TokenEndpoint)
)

//goland:noinspection GoNameStartsWithPackageName
type TokenEndpointConfigurer struct {
}

func newOAuth2TokenEndpointConfigurer() *TokenEndpointConfigurer {
	return &TokenEndpointConfigurer{
	}
}

func (c *TokenEndpointConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	// Verify
	f := feature.(*TokenFeature)
	if err := c.validate(f, ws); err != nil {
		return err
	}

	// prepare middlewares
	tokenMw := NewTokenEndpointMiddleware(func(opts *TokenEndpointOptions) {
		opts.Granter = auth.NewCompositeTokenGranter(f.granters...)
	})

	// install middlewares
	tokenMapping := middleware.NewBuilder("token endpoint").
		ApplyTo(matcher.RouteWithPattern(f.path, http.MethodPost)).
		Order(security.MWOrderOAuth2Endpoints).
		Use(tokenMw.TokenHandlerFunc())

	ws.Add(tokenMapping)

	// add dummy handler
	ws.Add(mapping.Post(f.path).HandlerFunc(security.NoopHandlerFunc()))

	return nil
}

func (c *TokenEndpointConfigurer) validate(f *TokenFeature, ws security.WebSecurity) error {
	if f.path == "" {
		return fmt.Errorf("token endpoint is not set")
	}

	if f.granters == nil || len(f.granters) == 0 {
		return fmt.Errorf("token granters is not set")
	}
	return nil
}



