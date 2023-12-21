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

package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	netutil "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/net"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// The OIDC RP initiated SLO is implemented by a set of handlers
// 1. ConditionalHandler
// 2. SuccessHandler
// 3. EntryPoint
// They work together according to the logic in security/logout/LogoutMiddleware
// First the ConditionalHandler is executed. The ConditionalHandler checks if all the conditions are met in order for us
//  to process the logout request. This means when the post_logout_redirect_uri is present, the id_token_hint is valid
//  and the redirect url can be verified according to the info presented by the id token. This happens in the ShouldLogout
//  method.
// If ShouldLogout is ok, Then the LogoutHandler chain continues, and the user is logged out.
//   after the logout handler chain finishes, the SuccessHandler is called, and the user is redirected according to the
//   post_logout_redirect_uri
// If Shouldlogout returns error, the logout process is also stopped, and the EntryPoint is called, which will direct user to an error page.

//nolint:gosec // not sure why linter think this is the case: "G101: Potential hardcoded credentials"
var ParameterRedirectUri = "post_logout_redirect_uri"
var ParameterIdTokenHint = "id_token_hint"
var ParameterState = "state"

type SuccessOptions func(opt *SuccessOption)

type SuccessOption struct {
	ClientStore         oauth2.OAuth2ClientStore
	WhitelabelErrorPath string
}

type OidcSuccessHandler struct {
	clientStore oauth2.OAuth2ClientStore
	fallback    security.AuthenticationErrorHandler
}

func (o *OidcSuccessHandler) Order() int {
	return order.Highest
}

func NewOidcSuccessHandler(opts ...SuccessOptions) *OidcSuccessHandler {
	opt := SuccessOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &OidcSuccessHandler{
		clientStore: opt.ClientStore,
		fallback:    redirect.NewRedirectWithURL(opt.WhitelabelErrorPath),
	}
}

func (o *OidcSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	redirectUri := r.FormValue(ParameterRedirectUri)
	if redirectUri == "" {
		// as OIDC success handler, we only care about this redirect
		return
	}

	state := r.FormValue(ParameterState)
	params := make(map[string]string)
	if state != "" {
		params[ParameterState] = state
	}
	redirectUri, err := netutil.AppendRedirectUrl(redirectUri, params)

	if err != nil {
		o.fallback.HandleAuthenticationError(c, r, rw, err)
		return
	}

	// since the corresponding logout handler already validated the logout request and the redirect uri, we just need to do the redirect.
	http.Redirect(rw, r, redirectUri, http.StatusFound)
	_, _ = rw.Write([]byte{})
}

type HandlerOptions func(opt *HandlerOption)

type HandlerOption struct {
	Dec         jwt.JwtDecoder
	Issuer      security.Issuer
	ClientStore oauth2.OAuth2ClientStore
}

type OidcLogoutHandler struct {
	dec         jwt.JwtDecoder
	issuer      security.Issuer
	clientStore oauth2.OAuth2ClientStore
}

func NewOidcLogoutHandler(opts ...HandlerOptions) *OidcLogoutHandler {
	opt := HandlerOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &OidcLogoutHandler{
		dec:         opt.Dec,
		issuer:      opt.Issuer,
		clientStore: opt.ClientStore,
	}
}

func (o *OidcLogoutHandler) Order() int {
	return order.Highest
}

func (o *OidcLogoutHandler) ShouldLogout(ctx context.Context, request *http.Request, writer http.ResponseWriter, authentication security.Authentication) error {
	switch request.Method {
	case http.MethodGet:
		fallthrough
	case http.MethodPost:
	case http.MethodPut:
		fallthrough
	case http.MethodDelete:
		fallthrough
	default:
		return ErrorOidcSloRp.WithMessage("unsupported http verb %v", request.Method)
	}

	//if logout request doesn't have this, we don't consider it a oidc logout request, and let other handle it.
	redirectUri := request.FormValue(ParameterRedirectUri)
	if redirectUri == "" {
		return nil
	}

	idTokenValue := request.FormValue(ParameterIdTokenHint)
	if strings.TrimSpace(idTokenValue) == "" {
		return ErrorOidcSloRp.WithMessage(`id token is required from parameter "%s"`, ParameterIdTokenHint)
	}

	claims, err := o.dec.Decode(ctx, idTokenValue)
	if err != nil {
		return ErrorOidcSloRp.WithMessage("id token invalid: %v", err)
	}

	iss := claims.Get(oauth2.ClaimIssuer)
	if iss != o.issuer.Identifier() {
		return ErrorOidcSloRp.WithMessage("id token is not issued by this auth server")
	}

	sub := claims.Get(oauth2.ClaimSubject)
	username, err := security.GetUsername(authentication)

	if err != nil {
		return ErrorOidcSloOp.WithMessage("Couldn't identify current session user")
	} else if sub != username {
		return ErrorOidcSloRp.WithMessage("logout request rejected because id token is not from the current session's user.")
	}

	clientId := claims.Get(oauth2.ClaimAudience).(string)
	client, err := auth.LoadAndValidateClientId(ctx, clientId, o.clientStore)
	if err != nil {
		return ErrorOidcSloOp.WithMessage("error loading client %s", clientId)
	}
	_, err = auth.ResolveRedirectUri(ctx, redirectUri, client)
	if err != nil {
		return ErrorOidcSloRp.WithMessage("redirect url %s is not registered by client %s", redirectUri, clientId)
	}

	r, err := url.Parse(redirectUri)
	if err != nil {
		return ErrorOidcSloRp.WithMessage("redirect url %s is not a valid url", redirectUri)
	} else {
		if r.RawQuery != "" {
			return ErrorOidcSloRp.WithMessage("redirect url %s should not contain query parameter", redirectUri)
		}
	}

	return nil
}

func (o *OidcLogoutHandler) HandleLogout(ctx context.Context, request *http.Request, writer http.ResponseWriter, authentication security.Authentication) error {
	//no op, because the default logout handler is sufficient (deleting the current session etc.)
	return nil
}

type EpOptions func(opt *EpOption)

type EpOption struct {
	WhitelabelErrorPath string
}

type OidcEntryPoint struct {
	fallback security.AuthenticationEntryPoint
}

func NewOidcEntryPoint(opts ...EpOptions) *OidcEntryPoint {
	opt := EpOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &OidcEntryPoint{
		fallback: redirect.NewRedirectWithURL(opt.WhitelabelErrorPath),
	}
}

func (o *OidcEntryPoint) Commence(ctx context.Context, request *http.Request, writer http.ResponseWriter, err error) {
	if !errors.Is(err, ErrorSubTypeOidcSlo) {
		return
	}
	switch {
	case errors.Is(err, ErrorOidcSloRp):
		fallthrough //currently we don't have any rp or op specific error handling requirements.
	case errors.Is(err, ErrorOidcSloOp):
		fallthrough
	default:
		o.fallback.Commence(ctx, request, writer, err)
	}
	return
}
