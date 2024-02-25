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

package basicauth

import (
    "context"
    "encoding/base64"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/passwd"
    "github.com/gin-gonic/gin"
    "net/http"
    "strconv"
    "strings"
)

//goland:noinspection GoNameStartsWithPackageName
type BasicAuthMiddleware struct {
	authenticator security.Authenticator
	successHandler security.AuthenticationSuccessHandler
}

func NewBasicAuthMiddleware(authenticator security.Authenticator, successHandler security.AuthenticationSuccessHandler) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{
		authenticator:  authenticator,
		successHandler: successHandler,
	}
}

func (basic *BasicAuthMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		header := ctx.GetHeader("Authorization")
		if header == "" {
			// Authorization header not available, bail
			return
		}
		if !strings.HasPrefix(header,"Basic ") {
			// Not basic auth, bail
			return
		}

		encoded := strings.TrimPrefix(header, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			basic.handleError(ctx, security.NewBadCredentialsError("invalid Authorization header"))
			return
		}

		pair := strings.SplitN(string(decoded), ":", 2)
		if len(pair) < 2 {
			basic.handleError(ctx, security.NewBadCredentialsError("invalid Authorization header"))
			return
		}

		before := security.Get(ctx)
		currentAuth, ok := before.(passwd.UsernamePasswordAuthentication)
		if ok && passwd.IsSamePrincipal(pair[0], currentAuth) {
			// already authenticated
			basic.handleSuccess(ctx, before, nil)
			return
		}

		candidate := passwd.UsernamePasswordPair{
			Username: pair[0],
			Password: pair[1],
		}
		// Search auth in the slice of allowed credentials
		auth, err := basic.authenticator.Authenticate(ctx, &candidate)
		if err != nil {
			basic.handleError(ctx, err)
			return
		}

		basic.handleSuccess(ctx, before, auth)
	}
}

func (basic *BasicAuthMiddleware) handleSuccess(c *gin.Context, before, new security.Authentication) {
	if new != nil {
		security.MustSet(c, new)
		basic.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, before, new)
	}
	c.Next()
}

func (basic *BasicAuthMiddleware) handleError(c *gin.Context, err error) {
	security.MustClear(c)
	_ = c.Error(err)
	c.Abort()
}

//goland:noinspection GoNameStartsWithPackageName
type BasicAuthEntryPoint struct {
	security.DefaultAuthenticationErrorHandler
}

func NewBasicAuthEntryPoint() *BasicAuthEntryPoint {
	return &BasicAuthEntryPoint{}
}

func (h *BasicAuthEntryPoint) Commence(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	writeBasicAuthChallenge(rw, err)
	h.DefaultAuthenticationErrorHandler.HandleAuthenticationError(c, r, rw, err)
}

//goland:noinspection GoNameStartsWithPackageName
type BasicAuthErrorHandler struct {
}

func NewBasicAuthErrorHandler() *BasicAuthErrorHandler {
	return &BasicAuthErrorHandler{}
}

func (h *BasicAuthErrorHandler) HandleAuthenticationError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	writeBasicAuthChallenge(rw, err)
}

func writeBasicAuthChallenge(rw http.ResponseWriter, err error) {
	realm := "Basic realm=" + strconv.Quote("Authorization Required")
	rw.Header().Set("WWW-Authenticate", realm)
}