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

package revoke

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
    "github.com/cisco-open/go-lanai/pkg/security/session"
    "net/http"
    "strings"
)

const (
	bearerTokenPrefix = "Bearer "
)

type HanlderOptions func(opt *HanlderOption)

type HanlderOption struct {
	Revoker auth.AccessRevoker
}

// TokenRevokingLogoutHandler
/**
 * GET method: used for logout by the session controlled clients. The client send user to this endpoint and the session
 * is invalidated. As a result, the tokens controlled by this session is invalidated (See the NfvClientDetails.useSessionTimeout
 * properties). It's also used by SSO logout (OIDC, and SAML GET Binding). In those case, the session is invalidated, and the
 * token controlled by the session is invalidated in the same way.
 *
 * POST method: used for logout by SSO logout (SAML POST Binding). The session is invalidated, and the token controlled
 * by the session is invalidated (same as the GET method).
 *
 * PUT/DELETE method: used for token revocation. Typically for service login or token revocation. We grab token
 * from header and revoke this only this token.
 *
 * @author Livan Du
 * Created on 2018-05-04
 */
type TokenRevokingLogoutHandler struct {
	revoker auth.AccessRevoker
}

func NewTokenRevokingLogoutHandler(opts ...HanlderOptions) *TokenRevokingLogoutHandler {
	opt := HanlderOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &TokenRevokingLogoutHandler{
		revoker: opt.Revoker,
	}
}

func (h TokenRevokingLogoutHandler) HandleLogout(ctx context.Context, r *http.Request, rw http.ResponseWriter, auth security.Authentication) error {
	switch r.Method {
	case http.MethodGet:
		fallthrough
	case http.MethodPost:
		return h.handleGetOrPost(ctx, auth)
	case http.MethodPut:
		fallthrough
	case http.MethodDelete:
		return h.handleDefault(ctx, r)
	}
	return nil
}

func (h TokenRevokingLogoutHandler) handleGetOrPost(ctx context.Context, auth security.Authentication) error {
	defer func() {
		security.MustClear(ctx)
		session.MustSet(ctx, nil)
	}()
	s := session.Get(ctx)
	if s == nil {
		logger.WithContext(ctx).Debugf("invalid use of GET/POST	 /logout endpoint. session is not found")
		return nil
	}

	if e := h.revoker.RevokeWithSessionId(ctx, s.GetID(), s.Name()); e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke tokens with session %s: %v", s.GetID(), e)
		return e
	}
	return nil
}

// In case of PUT, DELETE, PATCH etc, we don't clean authentication. Instead, we invalidate access token carried by header
func (h TokenRevokingLogoutHandler) handleDefault(ctx context.Context, r *http.Request) error {
	// grab token
	tokenValue, e := h.extractAccessToken(ctx, r)
	if e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke token: %v", e)
		return nil
	}

	if e := h.revoker.RevokeWithTokenValue(ctx, tokenValue, auth.RevokerHintAccessToken); e != nil {
		logger.WithContext(ctx).Warnf("unable to revoke token with value %s: %v", log.Capped(tokenValue, 20), e)
		return e
	}
	return nil
}

func (h TokenRevokingLogoutHandler) extractAccessToken(ctx context.Context, r *http.Request) (string, error) {
	// try header first
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToUpper(header), strings.ToUpper(bearerTokenPrefix)) {
		return header[len(bearerTokenPrefix):], nil
	}

	// then try param
	value := r.FormValue(oauth2.ParameterAccessToken)
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf(`access token is required either from "Authorization" header or parameter "%s"`, oauth2.ParameterAccessToken)
	}
	return value, nil
}
