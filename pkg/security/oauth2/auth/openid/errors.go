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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	errorutils "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/error"
	"errors"
	"net/http"
)

const (
	_ = iota
	// ErrorSubTypeCodeOidcSlo non-programming error that can occur during oidc RP initiated logout
	ErrorSubTypeCodeOidcSlo = security.ErrorTypeCodeOidc + iota<<errorutils.ErrorSubTypeOffset
)

const (
	_ = ErrorSubTypeCodeOidcSlo + iota
	ErrorCodeOidcSloRp
	ErrorCodeOidcSloOp
)

var (
	ErrorSubTypeOidcSlo = security.NewErrorSubType(ErrorSubTypeCodeOidcSlo, errors.New("error sub-type: oidc slo"))

	// ErrorOidcSloRp errors are displayed as an HTML page with status 400
	ErrorOidcSloRp = security.NewCodedError(ErrorCodeOidcSloRp, "SLO rp error")
	// ErrorOidcSloOp errors are displayed as an HTML page with status 500
	ErrorOidcSloOp = security.NewCodedError(ErrorCodeOidcSloOp, "SLO op error")
)

func newOpenIDExtendedError(oauth2Code string, value interface{}, causes []interface{}) error {
	return oauth2.NewOAuth2Error(oauth2.ErrorCodeOpenIDExt, value,
		oauth2Code, http.StatusBadRequest, causes...)
}

func NewOpenIDExtendedError(oauth2Code string, value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2Code, value, causes)
}

func NewInteractionRequiredError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationInteractionRequired, value, causes)
}

func NewLoginRequiredError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationLoginRequired, value, causes)
}

func NewAccountSelectionRequiredError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationAcctSelectRequired, value, causes)
}

func NewInvalidRequestURIError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationInvalidRequestURI, value, causes)
}

func NewInvalidRequestObjError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationInvalidRequestObj, value, causes)
}

func NewRequestNotSupportedError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationRequestUnsupported, value, causes)
}

func NewRequestURINotSupportedError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationRequestURIUnsupported, value, causes)
}

func NewRegistrationNotSupportedError(value interface{}, causes ...interface{}) error {
	return newOpenIDExtendedError(oauth2.ErrorTranslationRegistrationUnsupported, value, causes)
}
