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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

var DefaultProtectionMatcher = matcher.NotRequest(matcher.RequestWithMethods("GET", "HEAD", "TRACE", "OPTIONS"))
var DefaultIgnoreMatcher = matcher.NoneRequest()

type manager struct {
	tokenStore TokenStore
	requireProtection web.RequestMatcher
	ignoreProtection web.RequestMatcher
	parameterName string
	headerName string

}

func newManager(tokenStore TokenStore, csrfProtectionMatcher web.RequestMatcher, ignoreProtectionMatcher web.RequestMatcher) *manager {
	if csrfProtectionMatcher == nil {
		csrfProtectionMatcher = DefaultProtectionMatcher
	}

	if ignoreProtectionMatcher == nil {
		ignoreProtectionMatcher = DefaultIgnoreMatcher
	}

	return &manager{
		tokenStore: tokenStore,
		parameterName: security.CsrfParamName,
		headerName: security.CsrfHeaderName,
		requireProtection: csrfProtectionMatcher,
		ignoreProtection: ignoreProtectionMatcher,
	}
}

func (m *manager) CsrfHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedToken, err := m.tokenStore.LoadToken(c)

		// this means there's something wrong with reading the csrf token from storage - e.g. can't deserialize or can't access storage
		// this means we can't recover, so abort
		if err != nil {
			_ = c.Error(security.NewInternalError(err.Error()))
			c.Abort()
		}

		if expectedToken == nil {
			expectedToken = m.tokenStore.Generate(c, m.parameterName, m.headerName)
			err = m.tokenStore.SaveToken(c, expectedToken)
			if err != nil {
				_ = c.Error(security.NewInternalError(err.Error()))
				c.Abort()
			}
		}

		//This so that the templates knows what to render to
		//we don't depend on the value being stored in session to decouple it from the store implementation.
		c.Set(web.ContextKeyCsrf, expectedToken)

		matches, err := m.requireProtection.MatchesWithContext(c, c.Request)
		if err != nil {
			_ = c.Error(security.NewInternalError(err.Error()))
			c.Abort()
		}

		ignores, err := m.ignoreProtection.MatchesWithContext(c, c.Request)
		if err != nil {
			_ = c.Error(security.NewInternalError(err.Error()))
			c.Abort()
		}

		if matches && !ignores {
			actualToken := c.GetHeader(m.headerName)

			if actualToken == "" {
				actualToken, _ = c.GetPostForm(m.parameterName)
			}

			//both error case returns access denied, but the error message may be different
			if actualToken == "" {
				_ = c.Error(security.NewMissingCsrfTokenError("request is missing csrf token"))
				c.Abort()
			} else if actualToken != expectedToken.Value {
				_ = c.Error(security.NewInvalidCsrfTokenError("request has invalid csrf token"))
				c.Abort()
			}
		}
	}
}

type CsrfDeniedHandler struct {
	delegate security.AccessDeniedHandler
}

// implement order.Ordered
func (h *CsrfDeniedHandler) Order() int {
	return 0
}

// implement security.AccessDeniedHandler
func (h *CsrfDeniedHandler) HandleAccessDenied(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, security.ErrorSubTypeCsrf):
		h.delegate.HandleAccessDenied(c, r, rw, err)
	}
}

