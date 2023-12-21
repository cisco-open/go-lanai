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

package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
	"net/url"
)

/********************************************
	Helper functions for OAuth2 Redirects
 ********************************************/
func composeRedirectUrl(c context.Context, r *AuthorizeRequest, values map[string]string, useFragment bool) (string, error) {
	redirectUrl, ok := findRedirectUri(c, r)
	if !ok {
		return "", fmt.Errorf("redirect URI is unknown")
	}

	if state, ok := findRedirectState(c, r); ok {
		values[oauth2.ParameterState] = state
	}

	return appendRedirectUrl(redirectUrl, values)
}

func appendRedirectUrl(redirectUrl string, params map[string]string) (string, error) {
	loc, e := url.ParseRequestURI(redirectUrl)
	if e != nil || !loc.IsAbs() {
		return "", oauth2.NewInvalidRedirectUriError("invalid redirect_uri")
	}

	// TODO support fragments
	query := loc.Query()
	for k, v := range params {
		query.Add(k, v)
	}
	loc.RawQuery = query.Encode()

	return loc.String(), nil
}

func findRedirectUri(c context.Context, r *AuthorizeRequest) (string, bool) {
	value, ok := c.Value(oauth2.CtxKeyResolvedAuthorizeRedirect).(string)
	if !ok && r != nil {
		value = r.RedirectUri
		ok = true
	}
	return value, ok
}

func findRedirectState(c context.Context, r *AuthorizeRequest) (string, bool) {
	value, ok := c.Value(oauth2.CtxKeyResolvedAuthorizeState).(string)
	if !ok && r != nil {
		value = r.State
		ok = true
	}
	return value, ok
}



