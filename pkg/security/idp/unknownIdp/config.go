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

package unknownIdp

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/access"
    "github.com/cisco-open/go-lanai/pkg/security/config/authserver"
    "github.com/cisco-open/go-lanai/pkg/security/errorhandling"
    "github.com/cisco-open/go-lanai/pkg/security/idp"
    "github.com/cisco-open/go-lanai/pkg/security/redirect"
    "github.com/cisco-open/go-lanai/pkg/security/session"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
)

type NoIdpSecurityConfigurer struct {
}

func NewNoIdpSecurityConfigurer() *NoIdpSecurityConfigurer {
	return &NoIdpSecurityConfigurer{}
}

func (c *NoIdpSecurityConfigurer) Configure(ws security.WebSecurity, config *authserver.Configuration) {
	// For Authorize endpoint
	handler := redirect.NewRedirectWithURL(config.Endpoints.Error)
	condition := idp.RequestWithAuthenticationFlow(idp.UnknownIdp, config.IdpManager)

	ws.AndCondition(condition).
		With(session.New().SettingService(config.SessionSettingService)).
		With(access.New().
			Request(matcher.AnyRequest()).
			AllowIf(authenticatedWithMessage("Identity provider is not configured for this sub-domain")),
		).
		With(errorhandling.New().
			AccessDeniedHandler(handler),
		)
}

func authenticatedWithMessage(format string, v...interface{}) access.ControlFunc {
	return func(auth security.Authentication) (decision bool, reason error) {
		if auth.State() >= security.StateAuthenticated {
			return true, nil
		} else {
			return false, security.NewAccessDeniedError(fmt.Sprintf(format, v...))
		}
	}
}