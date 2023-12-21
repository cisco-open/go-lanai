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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"net/http"
	"net/url"
)

type SuccessOptions func(opt *SuccessOption)

type SuccessOption struct {
	ClientStore             oauth2.OAuth2ClientStore
	RedirectWhitelist       utils.StringSet
	WhitelabelErrorPath     string
	WhitelabelLoggedOutPath string
}

// TokenRevokeSuccessHandler implements security.AuthenticationSuccessHandler
type TokenRevokeSuccessHandler struct {
	clientStore           oauth2.OAuth2ClientStore
	whitelist             utils.StringSet
	fallback              security.AuthenticationErrorHandler
	defaultSuccessHandler security.AuthenticationSuccessHandler
}

func NewTokenRevokeSuccessHandler(opts ...SuccessOptions) *TokenRevokeSuccessHandler {
	opt := SuccessOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &TokenRevokeSuccessHandler{
		clientStore:           opt.ClientStore,
		fallback:              redirect.NewRedirectWithURL(opt.WhitelabelErrorPath),
		whitelist:             opt.RedirectWhitelist,
		defaultSuccessHandler: redirect.NewRedirectWithRelativePath(opt.WhitelabelLoggedOutPath, true),
	}
}

func (h TokenRevokeSuccessHandler) HandleAuthenticationSuccess(ctx context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	switch r.Method {
	case http.MethodGet:
		fallthrough
	case http.MethodPost:
		h.redirect(ctx, r, rw, from, to)
	case http.MethodPut:
		fallthrough
	case http.MethodDelete:
		fallthrough
	default:
		h.status(ctx, rw)
	}
}

func (h TokenRevokeSuccessHandler) redirect(ctx context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	// Note: we don't have error handling alternatives (except for panic)
	redirectUri := r.FormValue(oauth2.ParameterRedirectUri)
	if redirectUri == "" {
		h.defaultSuccessHandler.HandleAuthenticationSuccess(ctx, r, rw, from, to)
		return
	}

	clientId := r.FormValue(oauth2.ParameterClientId)
	client, e := auth.LoadAndValidateClientId(ctx, clientId, h.clientStore)
	if e != nil {
		h.fallback.HandleAuthenticationError(ctx, r, rw, e)
		return
	}

	resolved, e := auth.ResolveRedirectUri(ctx, redirectUri, client)
	if e != nil {
		// try resolve from whitelist
		if !h.isWhitelisted(ctx, redirectUri) {
			h.fallback.HandleAuthenticationError(ctx, r, rw, e)
			return
		}
		resolved = redirectUri
	}

	h.doRedirect(ctx, r, rw, resolved)
}

func (h TokenRevokeSuccessHandler) doRedirect(ctx context.Context, r *http.Request, rw http.ResponseWriter, redirectUri string) {
	redirectUrl := h.appendWarnings(ctx, redirectUri)
	http.Redirect(rw, r, redirectUrl, http.StatusFound)
	_, _ = rw.Write([]byte{})
}

// In case of PUT, DELETE, PATCH etc, we don't clean authentication. Instead, we invalidate access token carried by header
func (h TokenRevokeSuccessHandler) status(_ context.Context, rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte{})
}

func (h TokenRevokeSuccessHandler) isWhitelisted(_ context.Context, redirect string) bool {
	for pattern, _ := range h.whitelist {
		matcher, e := auth.NewWildcardUrlMatcher(pattern)
		if e != nil {
			continue
		}
		if matches, e := matcher.Matches(redirect); e == nil && matches {
			return true
		}
	}
	return false
}

func (h TokenRevokeSuccessHandler) appendWarnings(ctx context.Context, redirect string) string {
	warnings := logout.GetWarnings(ctx)
	if len(warnings) == 0 {
		return redirect
	}

	redirectUrl, e := url.Parse(redirect)
	if e != nil {
		return redirect
	}

	q := redirectUrl.Query()
	for _, w := range warnings {
		q.Add("warning", w.Error())
	}
	redirectUrl.RawQuery = q.Encode()
	return redirectUrl.String()
}
