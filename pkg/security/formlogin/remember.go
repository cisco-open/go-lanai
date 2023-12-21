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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"math"
	"net/http"
	"net/url"
	"time"
)

const (
	detailsKeyShouldRememberUsername = "RememberUsername"
)

type rememberMeOptions func(h *RememberUsernameSuccessHandler)

type RememberUsernameSuccessHandler struct {
	contextPath    string
	rememberParam  string
	cookieDomain   string
	cookieSecured  bool
	cookieHttpOnly bool
	cookieMaxAge   time.Duration
}

func newRememberUsernameSuccessHandler(opts ...rememberMeOptions) *RememberUsernameSuccessHandler {
	h := RememberUsernameSuccessHandler{
		contextPath: "/",
	}
	for _, fn := range opts {
		fn(&h)
	}
	return &h
}

func (h *RememberUsernameSuccessHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, _, to security.Authentication) {
	details, ok := to.Details().(map[string]interface{})
	if !ok {
		details = map[string]interface{}{}
	}

	// set remember-me decision to auth's details if request has such parameter
	remember := r.PostForm.Get(h.rememberParam)
	if remember != "" {
		details[detailsKeyShouldRememberUsername] = true
	}

	// auth process not finished yet, bail
	if to.State() < security.StateAuthenticated {
		return
	}

	// read remember-me decision from auth
	if doRemember, ok := details[detailsKeyShouldRememberUsername].(bool); !ok || !doRemember {
		// cleanup session
		h.clear(c, rw)
		return
	}

	// remember username
	switch to.Principal().(type) {
	case security.Account:
		h.save(to.Principal().(security.Account).Username(), c, rw)
	case string:
		h.save(to.Principal().(string), c, rw)
	}
}

func (h *RememberUsernameSuccessHandler) save(username string, _ context.Context, rw http.ResponseWriter) {
	cookie := h.newCookie(CookieKeyRememberedUsername, username, h.cookieMaxAge)
	http.SetCookie(rw, cookie)
}

func (h *RememberUsernameSuccessHandler) clear(_ context.Context, rw http.ResponseWriter) {
	cookie := h.newCookie(CookieKeyRememberedUsername, "", -1)
	http.SetCookie(rw, cookie)
}

func (h *RememberUsernameSuccessHandler) newCookie(name, value string, maxAge time.Duration) *http.Cookie {

	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     h.contextPath,
		Domain:   h.cookieDomain,
		MaxAge:   int(math.Round(maxAge.Seconds())),
		Expires:  calculateCookieExpires(maxAge),
		Secure:   h.cookieSecured,
		HttpOnly: h.cookieHttpOnly,
		SameSite: http.SameSiteStrictMode, //The remember me cookie should not be used cross site (unlike the session cookie which need to work cross site for sso)
	}

	return cookie
}

func calculateCookieExpires(maxAge time.Duration) time.Time {
	if maxAge == 0 {
		return time.Time{}
	}

	return time.Now().Add(maxAge)
}
