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
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "net/http"
)

var (
	errorInvalidRedirectUri = oauth2.NewInvalidRedirectUriError("")
)

// OAuth2ErrorHandler implements security.ErrorHandler
// It's responsible to handle all oauth2 errors
type OAuth2ErrorHandler struct {}

func NewOAuth2ErrorHandler() *OAuth2ErrorHandler {
	return &OAuth2ErrorHandler{}
}

// HandleError implements security.ErrorHandler
func (h *OAuth2ErrorHandler) HandleError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	h.handleError(c, r, rw, err)
}

func (h *OAuth2ErrorHandler) handleError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	//nolint:errorlint
	switch oe, ok := err.(oauth2.OAuth2ErrorTranslator); {
	case ok && errors.Is(err, errorInvalidRedirectUri):
		writeOAuth2Error(c, r, rw, oe)
	case ok && errors.Is(err, oauth2.ErrorSubTypeOAuth2Internal):
		fallthrough
	case ok && errors.Is(err, oauth2.ErrorTypeOAuth2):
		// use redirect uri, fallback to standard error response
		tryWriteErrorAsRedirect(c, r, rw, oe)
	// No default, give other error handler chance to handle
	}
}

func writeOAuth2Error(c context.Context, r *http.Request, rw http.ResponseWriter, err oauth2.OAuth2ErrorTranslator) {
	challenge := ""
	sc := err.TranslateStatusCode()
	if sc == http.StatusUnauthorized || sc == http.StatusForbidden {
		challenge = fmt.Sprintf("%s %s", "Bearer", err.Error())
	}
	writeAdditionalHeader(c, r, rw, challenge)
	switch {
	case errors.Is(err, errorInvalidRedirectUri):
		security.WriteError(c, r, rw, sc, err)
	default:
		security.WriteErrorAsJson(c, rw, sc, err)
	}
}

func writeAdditionalHeader(_ context.Context, _ *http.Request, rw http.ResponseWriter, challenge string) {
	if security.IsResponseWritten(rw) {
		return
	}

	rw.Header().Add("Cache-Control", "no-store")
	rw.Header().Add("Pragma", "no-cache")

	if challenge != "" {
		rw.Header().Set("WWW-Authenticate", challenge)
	}
}

// given err have to be oauth2.OAuth2ErrorTranslator
func tryWriteErrorAsRedirect(c context.Context, r *http.Request, rw http.ResponseWriter, err oauth2.OAuth2ErrorTranslator) {
	if security.IsResponseWritten(rw) {
		return
	}

	params := map[string]string{}
	params[oauth2.ParameterError] = err.TranslateErrorCode()
	params[oauth2.ParameterErrorDescription] = err.Error()

	// TODO support fragment
	ar := findAuthorizeRequest(c, r)
	redirectUrl, e := composeRedirectUrl(c, ar, params, false)
	if e != nil {
		// fallback to default
		writeOAuth2Error(c, r, rw, err)
		return
	}
	http.Redirect(rw, r, redirectUrl, http.StatusFound)
	_, _ = rw.Write([]byte{})
}

func findAuthorizeRequest(c context.Context, _ *http.Request) *AuthorizeRequest {
	if ar, ok := c.Value(oauth2.CtxKeyValidatedAuthorizeRequest).(*AuthorizeRequest); ok {
		return ar
	}

	if ar, ok := c.Value(oauth2.CtxKeyReceivedAuthorizeRequest).(*AuthorizeRequest); ok {
		return ar
	}

	return nil
}



