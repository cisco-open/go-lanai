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

package formlogin

import (
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/passwd"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/gin-gonic/gin"
)

type FormAuthenticationMiddleware struct {
	authenticator  security.Authenticator
	successHandler security.AuthenticationSuccessHandler
	usernameParam  string
	passwordParam  string
}

type FormAuthMWOptionsFunc func(*FormAuthMWOptions)

type FormAuthMWOptions struct {
	Authenticator  security.Authenticator
	SuccessHandler security.AuthenticationSuccessHandler
	UsernameParam  string
	PasswordParam  string
}

func NewFormAuthenticationMiddleware(optionFuncs... FormAuthMWOptionsFunc) *FormAuthenticationMiddleware {
	opts := FormAuthMWOptions{}
	for _, optFunc := range optionFuncs {
		if optFunc != nil {
			optFunc(&opts)
		}
	}
	return &FormAuthenticationMiddleware{
		authenticator: opts.Authenticator,
		successHandler: opts.SuccessHandler,
		usernameParam: opts.UsernameParam,
		passwordParam: opts.PasswordParam,
	}
}

func (mw *FormAuthenticationMiddleware) LoginProcessHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		username := ctx.PostFormArray(mw.usernameParam)
		if len(username) == 0 {
			username = []string{""}
		}

		password := ctx.PostFormArray(mw.passwordParam)
		if len(password) == 0 {
			password = []string{""}
		}

		before := security.Get(ctx)
		currentAuth, ok := before.(passwd.UsernamePasswordAuthentication)

		//nolint:staticcheck // empty block for document purpose
		if ok && passwd.IsSamePrincipal(username[0], currentAuth) {
			// We currently allow re-authenticate without logout.
			// If we don't want to allow it, we need to figure out how to error out without clearing the authentication.
			// Note: currently, clearing authentication happens in error handling middleware on all SecurityAuthenticationError
		}

		candidate := passwd.UsernamePasswordPair{
			Username: username[0],
			Password: password[0],
			EnforceMFA: passwd.MFAModeOptional,
		}
		// Authenticate
		auth, err := mw.authenticator.Authenticate(ctx, &candidate)
		if err != nil {
			mw.handleError(ctx, err, &candidate)
			return
		}
		mw.handleSuccess(ctx, before, auth)
	}
}

func (mw *FormAuthenticationMiddleware) handleSuccess(c *gin.Context, before, new security.Authentication) {
	if new != nil {
		security.MustSet(c, new)
	}
	mw.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, before, new)
	if c.Writer.Written() {
		c.Abort()
	}
}

func (mw *FormAuthenticationMiddleware) handleError(c *gin.Context, err error, candidate security.Candidate) {
	security.MustClear(c)
	if candidate != nil {
		s := session.Get(c)
		if s != nil {
			s.AddFlash(candidate.Principal(), mw.usernameParam)
		}
	}
	_ = c.Error(err)
	c.Abort()
}
