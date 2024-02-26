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

package oauth2

import (
    "bytes"
    "encoding/gob"
    "encoding/json"
    "errors"
    "github.com/cisco-open/go-lanai/pkg/security"
    errorutils "github.com/cisco-open/go-lanai/pkg/utils/error"
    "net/http"
)

// All "SubType" values are used as mask
// sub types of security.ErrorTypeCodeOAuth2
const (
	_                              = iota
	ErrorSubTypeCodeOAuth2Internal = security.ErrorTypeCodeOAuth2 + iota<<errorutils.ErrorSubTypeOffset
	ErrorSubTypeCodeOAuth2ClientAuth
	ErrorSubTypeCodeOAuth2Authorize
	ErrorSubTypeCodeOAuth2Grant
	ErrorSubTypeCodeOAuth2Res
)

// ErrorSubTypeCodeOAuth2Internal
const (
	_ = ErrorSubTypeCodeOAuth2Internal + iota
	ErrorCodeOAuth2InternalGeneral
)

// ErrorSubTypeCodeOAuth2ClientAuth
const (
	_ = ErrorSubTypeCodeOAuth2ClientAuth + iota
	ErrorCodeClientNotFound
	ErrorCodeInvalidClient
)

// ErrorSubTypeCodeOAuth2Authorize
const (
	_ = ErrorSubTypeCodeOAuth2Authorize + iota
	ErrorCodeInvalidAuthorizeRequest
	ErrorCodeInvalidResponseType
	ErrorCodeInvalidRedirectUri
	ErrorCodeAccessRejected
	ErrorCodeOpenIDExt
)

// ErrorSubTypeCodeOAuth2Grant
const (
	_ = ErrorSubTypeCodeOAuth2Grant + iota
	ErrorCodeGranterNotAvailable
	ErrorCodeUnauthorizedClient // grant type is not allowed for client
	ErrorCodeInvalidTokenRequest
	ErrorCodeInvalidGrant
	ErrorCodeInvalidScope
	ErrorCodeUnsupportedTokenType
	ErrorCodeGeneric
)

// ErrorSubTypeCodeOAuth2Res
const (
	_ = ErrorSubTypeCodeOAuth2Res + iota
	ErrorCodeInvalidAccessToken
	ErrorCodeInsufficientScope
	ErrorCodeResourceServerGeneral // this should only be used for error deserialization
)

// ErrorTypes, can be used in errors.Is
//goland:noinspection GoUnusedGlobalVariable
var (
	ErrorTypeOAuth2 = security.NewErrorType(security.ErrorTypeCodeOAuth2, errors.New("error type: oauth2"))

	ErrorSubTypeOAuth2Internal   = security.NewErrorSubType(ErrorSubTypeCodeOAuth2Internal, errors.New("error sub-type: internal"))
	ErrorSubTypeOAuth2ClientAuth = security.NewErrorSubType(ErrorSubTypeCodeOAuth2ClientAuth, errors.New("error sub-type: oauth2 client auth"))
	ErrorSubTypeOAuth2Authorize  = security.NewErrorSubType(ErrorSubTypeCodeOAuth2Authorize, errors.New("error sub-type: oauth2 auth"))
	ErrorSubTypeOAuth2Grant      = security.NewErrorSubType(ErrorSubTypeCodeOAuth2Grant, errors.New("error sub-type: oauth2 grant"))
	ErrorSubTypeOAuth2Res        = security.NewErrorSubType(ErrorSubTypeCodeOAuth2Res, errors.New("error sub-type: oauth2 resource"))
)

/************************
	Error EC
*************************/

//goland:noinspection GoCommentStart
const (
	// https://tools.ietf.org/html/rfc6749#section-4.1.2.1
	ErrorTranslationInvalidRequest      = "invalid_request"
	ErrorTranslationUnauthorizedClient  = "unauthorized_client"
	ErrorTranslationAccessDenied        = "access_denied"
	ErrorTranslationInvalidResponseType = "unsupported_response_type"
	ErrorTranslationInvalidScope        = "invalid_scope"
	ErrorTranslationInternal            = "server_error"
	ErrorTranslationInternalNA          = "temporarily_unavailable"

	// https://tools.ietf.org/html/rfc6749#section-5.2
	ErrorTranslationInvalidClient     = "invalid_client"
	ErrorTranslationInvalidGrant      = "invalid_grant"
	ErrorTranslationGrantNotSupported = "unsupported_grant_type"

	// commonly used (no RFC reference for now)
	ErrorTranslationInsufficientScope = "insufficient_scope"
	ErrorTranslationInvalidToken      = "invalid_token"
	ErrorTranslationRedirectMismatch  = "redirect_uri_mismatch"

	// https://tools.ietf.org/html/rfc7009#section-4.1.1
	ErrorTranslationUnsupportedTokenType = "unsupported_token_type"

	// https://openid.net/specs/openid-connect-core-1_0.html#AuthError
	ErrorTranslationInteractionRequired     = "interaction_required"
	ErrorTranslationLoginRequired           = "login_required"
	ErrorTranslationAcctSelectRequired      = "account_selection_required"
	ErrorTranslationConsentRequired         = "consent_required"
	ErrorTranslationInvalidRequestURI       = "invalid_request_uri"
	ErrorTranslationInvalidRequestObj       = "invalid_request_object"
	ErrorTranslationRequestUnsupported      = "request_not_supported"
	ErrorTranslationRequestURIUnsupported   = "request_uri_not_supported"
	ErrorTranslationRegistrationUnsupported = "registration_not_supported"
	//ErrorTranslation = ""
)

/************************
	Extensions
*************************/

//goland:noinspection GoNameStartsWithPackageName
type OAuth2ErrorTranslator interface {
	error
	TranslateErrorCode() string
	TranslateStatusCode() int
}

// OAuth2Error extends security.CodedError, and implements:
//	- OAuth2ErrorTranslator
//  - json.Marshaler
// 	- json.Unmarshaler
// 	- web.Headerer
// 	- web.StatusCoder
//  - encoding.BinaryMarshaler
//  - encoding.BinaryUnmarshaler
//goland:noinspection GoNameStartsWithPackageName
type OAuth2Error struct {
	security.CodedError
	EC string // oauth error code
	SC int    // status code
}

func (e *OAuth2Error) StatusCode() int {
	return e.SC
}

func (e *OAuth2Error) Headers() http.Header {
	header := http.Header{}
	header.Add("Cache-Control", "no-store")
	header.Add("Pragma", "no-cache")
	return header
}

func (e *OAuth2Error) TranslateErrorCode() string {
	return e.EC
}

func (e *OAuth2Error) TranslateStatusCode() int {
	return e.SC
}

// MarshalJSON implements json.Marshaler
func (e *OAuth2Error) MarshalJSON() ([]byte, error) {
	data := map[string]string{
		ParameterError:            e.EC,
		ParameterErrorDescription: e.Error(),
	}
	return json.Marshal(data)
}

// UnmarshalJSON implements json.Unmarshaler
// Note: JSON doesn't include internal code error. So reconstruct error from JSON is not possible.
//       Unmarshaler can only be used for opaque token checking HTTP call
func (e *OAuth2Error) UnmarshalJSON(data []byte) error {
	values := map[string]string{}
	if e := json.Unmarshal(data, &values); e != nil {
		return e
	}

	e.EC = values[ParameterError]

	desc := values[ParameterErrorDescription]
	e.CodedError = *security.NewCodedError(ErrorCodeResourceServerGeneral, desc)
	return nil
}

type oauth2ErrorCarrier struct {
	CodedError security.CodedError
	EC         string // oauth error code
	SC         int    // status code
}

// MarshalBinary implements encoding.BinaryMarshaler interface
func (e OAuth2Error) MarshalBinary() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buffer)
	carrier := oauth2ErrorCarrier{
		CodedError: e.CodedError,
		EC:         e.EC,
		SC:         e.SC,
	}
	if e := encoder.Encode(&carrier); e != nil {
		return nil, e
	}
	return buffer.Bytes(), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface
func (e *OAuth2Error) UnmarshalBinary(data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	carrier := oauth2ErrorCarrier{}
	if e := decoder.Decode(&carrier); e != nil {
		return e
	}
	*e = OAuth2Error{
		CodedError: carrier.CodedError,
		EC:         carrier.EC,
		SC:         carrier.SC,
	}
	return nil
}

/************************
	Constructors
*************************/

func NewOAuth2Error(code int64, e interface{}, oauth2Code string, sc int, causes ...interface{}) *OAuth2Error {
	embedded := security.NewCodedError(code, e, causes...)
	return &OAuth2Error{
		CodedError: *embedded,
		EC:         oauth2Code,
		SC:         sc,
	}
}

/* OAuth2Internal family */

func NewInternalError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeOAuth2InternalGeneral, value,
		ErrorTranslationInternal, http.StatusBadRequest,
		causes...)
}

func NewInternalUnavailableError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeOAuth2InternalGeneral, value,
		ErrorTranslationInternalNA, http.StatusBadRequest,
		causes...)
}

/* OAuth2Auth family */

func NewGranterNotAvailableError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeGranterNotAvailable, value,
		ErrorTranslationGrantNotSupported, http.StatusBadRequest,
		causes...)
}

func NewInvalidTokenRequestError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeInvalidTokenRequest, value,
		ErrorTranslationInvalidRequest, http.StatusBadRequest,
		causes...)
}

func NewInvalidClientError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeInvalidClient, value,
		ErrorTranslationInvalidClient, http.StatusUnauthorized,
		causes...)
}

func NewClientNotFoundError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeClientNotFound, value,
		ErrorTranslationInvalidClient, http.StatusUnauthorized,
		causes...)
}

func NewUnauthorizedClientError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeUnauthorizedClient, value,
		ErrorTranslationUnauthorizedClient, http.StatusBadRequest,
		causes...)
}

func NewInvalidGrantError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeInvalidGrant, value,
		ErrorTranslationInvalidGrant, http.StatusBadRequest,
		causes...)
}

func NewInvalidScopeError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeInvalidScope, value,
		ErrorTranslationInvalidScope, http.StatusBadRequest,
		causes...)
}

func NewUnsupportedTokenTypeError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeUnsupportedTokenType, value,
		ErrorTranslationUnsupportedTokenType, http.StatusBadRequest,
		causes...)
}

func NewGenericError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeGeneric, value,
		ErrorTranslationInvalidRequest, http.StatusBadRequest,
		causes...)
}

func NewInvalidAuthorizeRequestError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeInvalidAuthorizeRequest, value,
		ErrorTranslationInvalidRequest, http.StatusBadRequest,
		causes...)
}

func NewInvalidRedirectUriError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeInvalidRedirectUri, value,
		ErrorTranslationRedirectMismatch, http.StatusBadRequest,
		causes...)
}

func NewInvalidResponseTypeError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeInvalidResponseType, value,
		ErrorTranslationInvalidResponseType, http.StatusBadRequest,
		causes...)
}

func NewAccessRejectedError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeAccessRejected, value,
		ErrorTranslationAccessDenied, http.StatusBadRequest,
		causes...)
}

/* OAuth2Res family */

func NewInvalidAccessTokenError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeInvalidAccessToken, value,
		ErrorTranslationInvalidToken, http.StatusUnauthorized,
		causes...)
}

func NewInsufficientScopeError(value interface{}, causes ...interface{}) error {
	return NewOAuth2Error(ErrorCodeInsufficientScope, value,
		ErrorTranslationInsufficientScope, http.StatusForbidden,
		causes...)
}
