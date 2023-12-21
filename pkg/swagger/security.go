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

package swagger

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

type swaggerSecurityConfigurer struct {
}

func (c *swaggerSecurityConfigurer) Configure(ws security.WebSecurity) {
	// DSL style example
	// for REST API
	ws.Route(matcher.RouteWithPattern("/v2/api-docs").Or(matcher.RouteWithPattern("/v3/api-docs"))).
		With(tokenauth.New()).
		With(access.New().
			Request(matcher.AnyRequest()).AllowIf(swaggerSpecAccessControl),
		).
		With(errorhandling.New())
}

func swaggerSpecAccessControl(auth security.Authentication) (decision bool, reason error) {
	oa, ok := auth.(oauth2.Authentication)
	if !ok {
		return false, security.NewInsufficientAuthError("expected token authentication")
	}

	if oa.UserAuthentication() == nil {
		return false, security.NewInsufficientAuthError("expected oauth user authentication")
	}

	if !(oa.OAuth2Request().Approved() && oa.OAuth2Request().Scopes().Has("read") && oa.OAuth2Request().Scopes().Has("write")) {
		return false, security.NewInsufficientAuthError("expected read and write scope")
	}

	//and must be authenticated
	return access.Authenticated(auth)
}
