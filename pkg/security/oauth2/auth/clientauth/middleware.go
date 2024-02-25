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

package clientauth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Middleware struct {
	authenticator  security.Authenticator
	successHandler security.AuthenticationSuccessHandler
}

type MWOptions func(*MWOption)

type MWOption struct {
	Authenticator  security.Authenticator
	SuccessHandler security.AuthenticationSuccessHandler
}

func NewClientAuthMiddleware(opts...MWOptions) *Middleware {
	opt := MWOption{}

	for _, optFunc := range opts {
		if optFunc != nil {
			optFunc(&opt)
		}
	}
	return &Middleware{
		authenticator:  opt.Authenticator,
		successHandler: opt.SuccessHandler,
	}
}

func (mw *Middleware) ClientAuthFormHandlerFunc() web.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if e := r.ParseForm(); e != nil {
			return
		}

		_, hasClientId := r.Form[oauth2.ParameterClientId]
		if !hasClientId {
			return
		}

		clientId := r.Form.Get(oauth2.ParameterClientId)

		// form client auth should be placed after basic auth.
		// if already authenticated by basic auth and pricipal matches, we don't need to do anything here
		// if authenticated but pricipal doesn't match, it's an error
		before := security.Get(r.Context())
		currentAuth, ok := before.(passwd.UsernamePasswordAuthentication)
		switch {
		case ok && passwd.IsSamePrincipal(clientId, currentAuth):
			return
		case ok:
			mw.handleError(r.Context(), oauth2.NewInvalidClientError("client_id parameter and Authorization header doesn't match"))
		}

		secret := r.PostForm.Get(oauth2.ParameterClientSecret)
		candidate := passwd.UsernamePasswordPair{
			Username: clientId,
			Password: secret,
			EnforceMFA: passwd.MFAModeSkip,
		}
		// Authenticate
		auth, err := mw.authenticator.Authenticate(r.Context(), &candidate)
		if err != nil {
			mw.handleError(r.Context(), err)
			return
		}
		mw.handleSuccess(r.Context(), r, rw, before, auth)
	}
}

func (mw *Middleware) ErrorTranslationHandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// find first authentication error and translate it
		for _, e := range c.Errors {
			switch {
			case errors.Is(e.Err, security.ErrorTypeAuthentication):
				e.Err = oauth2.NewInvalidClientError("client authentication failed", e.Err)
			}
		}
	}
}

func (mw *Middleware) handleSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, before, new security.Authentication) {
	gc := web.GinContext(c)
	if new != nil {
		security.MustSet(c, new)
		mw.successHandler.HandleAuthenticationSuccess(c, r, rw, before, new)
	}
	gc.Next()
}

//nolint:contextcheck
func (mw *Middleware) handleError(c context.Context, err error) {
	gc := web.GinContext(c)
	security.MustClear(gc)
	_ = gc.Error(err)
	gc.Abort()
}