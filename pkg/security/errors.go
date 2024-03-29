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

package security

import (
    "errors"
    errorutils "github.com/cisco-open/go-lanai/pkg/utils/error"
)

const (
	// security reserved
	Reserved = 11 << errorutils.ReservedOffset
)

// All "Type" values are used as mask
const (
	_                           = iota
	ErrorTypeCodeAuthentication = Reserved + iota<<errorutils.ErrorTypeOffset
	ErrorTypeCodeAccessControl
	ErrorTypeCodeInternal
	ErrorTypeCodeOAuth2
	ErrorTypeCodeSaml
	ErrorTypeCodeOidc
	ErrorTypeCodeTenancy
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeAuthentication
const (
	_                        = iota
	ErrorSubTypeCodeInternal = ErrorTypeCodeAuthentication + iota<<errorutils.ErrorSubTypeOffset
	ErrorSubTypeCodeUsernamePasswordAuth
	ErrorSubTypeCodeExternalSamlAuth
	ErrorSubTypeCodeAuthWarning
)

// ErrorSubTypeCodeInternal
const (
	_                                  = iota
	ErrorCodeAuthenticatorNotAvailable = ErrorSubTypeCodeInternal + iota
)

// ErrorSubTypeCodeUsernamePasswordAuth
const (
	_                         = iota
	ErrorCodeUsernameNotFound = ErrorSubTypeCodeUsernamePasswordAuth + iota
	ErrorCodeBadCredentials
	ErrorCodeCredentialsExpired
	ErrorCodeMaxAttemptsReached
	ErrorCodeAccountStatus
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeAccessControl
const (
	_                            = iota
	ErrorSubTypeCodeAccessDenied = ErrorTypeCodeAccessControl + iota<<errorutils.ErrorSubTypeOffset
	ErrorSubTypeCodeInsufficientAuth
	ErrorSubTypeCodeCsrf
)

// All "SubType" values are used as mask
// sub types of ErrorTypeCodeTenancy
const (
	_                             = iota
	ErrorSubTypeCodeTenantInvalid = ErrorTypeCodeTenancy + iota<<errorutils.ErrorSubTypeOffset
	ErrorSubTypeCodeTenantAccessDenied
)

const (
	_                         = iota
	ErrorCodeMissingCsrfToken = ErrorSubTypeCodeCsrf + iota
	ErrorCodeInvalidCsrfToken
)

// ErrorTypes, can be used in errors.Is
var (
	ErrorTypeSecurity       = NewErrorCategory(Reserved, errors.New("error type: security"))
	ErrorTypeAuthentication = NewErrorType(ErrorTypeCodeAuthentication, errors.New("error type: authentication"))
	ErrorTypeAccessControl  = NewErrorType(ErrorTypeCodeAccessControl, errors.New("error type: access control"))
	ErrorTypeInternal       = NewErrorType(ErrorTypeCodeInternal, errors.New("error type: internal"))
	ErrorTypeSaml           = NewErrorType(ErrorTypeCodeSaml, errors.New("error type: saml"))
	ErrorTypeOidc           = NewErrorType(ErrorTypeCodeOidc, errors.New("error type: oidc"))

	ErrorSubTypeInternalError        = NewErrorSubType(ErrorSubTypeCodeInternal, errors.New("error sub-type: internal"))
	ErrorSubTypeUsernamePasswordAuth = NewErrorSubType(ErrorSubTypeCodeUsernamePasswordAuth, errors.New("error sub-type: internal"))
	ErrorSubTypeExternalSamlAuth     = NewErrorSubType(ErrorSubTypeCodeExternalSamlAuth, errors.New("error sub-type: external saml"))
	ErrorSubTypeAuthWarning          = NewErrorSubType(ErrorSubTypeCodeAuthWarning, errors.New("error sub-type: auth warning"))

	ErrorSubTypeAccessDenied     = NewErrorSubType(ErrorSubTypeCodeAccessDenied, errors.New("error sub-type: access denied"))
	ErrorSubTypeInsufficientAuth = NewErrorSubType(ErrorSubTypeCodeInsufficientAuth, errors.New("error sub-type: insufficient auth"))
	ErrorSubTypeCsrf             = NewErrorSubType(ErrorSubTypeCodeCsrf, errors.New("error sub-type: csrf"))
)

// Concrete error, can be used in errors.Is for exact match
var (
	ErrorInvalidTenantId    = NewCodedError(ErrorSubTypeCodeTenantInvalid, "Invalid tenant Id")
	ErrorTenantAccessDenied = NewCodedError(ErrorSubTypeCodeTenantAccessDenied, "No Access to the tenant")
)

func init() {
	errorutils.Reserve(ErrorTypeSecurity)
}

// CodedError implements errorutils.ErrorCoder, errorutils.ComparableErrorCoder, errorutils.NestedError
type CodedError struct {
	errorutils.CodedError
}

/************************
	Constructors
*************************/

// NewCodedError creates concrete error. it cannot be used as ErrorType or ErrorSubType comparison
// supported item are string, error, fmt.Stringer
func NewCodedError(code int64, e interface{}, causes ...interface{}) *CodedError {
	return &CodedError{
		CodedError: *errorutils.NewCodedError(code, e, causes...),
	}
}

func NewErrorCategory(code int64, e error) *CodedError {
	return &CodedError{
		CodedError: *errorutils.NewErrorCategory(code, e),
	}
}

func NewErrorType(code int64, e error) error {
	return errorutils.NewErrorType(code, e)
}

func NewErrorSubType(code int64, e error) error {
	return errorutils.NewErrorSubType(code, e)
}

/* InternalError family */

func NewInternalError(text string, causes ...interface{}) error {
	return NewCodedError(ErrorTypeCodeInternal, errors.New(text), causes...)
}

/* AuthenticationError family */

func NewAuthenticationError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorTypeCodeAuthentication, value, causes...)
}

func NewInternalAuthenticationError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorSubTypeCodeInternal, value, causes...)
}

func NewAuthenticationWarningError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorSubTypeCodeAuthWarning, value, causes...)
}

func NewAuthenticatorNotAvailableError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorCodeAuthenticatorNotAvailable, value, causes...)
}

func NewExternalSamlAuthenticationError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorSubTypeCodeExternalSamlAuth, value, causes...)
}

func NewUsernameNotFoundError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorCodeUsernameNotFound, value, causes...)
}

func NewBadCredentialsError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorCodeBadCredentials, value, causes...)
}

func NewCredentialsExpiredError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorCodeCredentialsExpired, value, causes...)
}

func NewMaxAttemptsReachedError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorCodeMaxAttemptsReached, value, causes...)
}

func NewAccountStatusError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorCodeAccountStatus, value, causes...)
}

/* AccessControlError family */

func NewAccessControlError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorTypeCodeAccessControl, value, causes...)
}

func NewAccessDeniedError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorSubTypeCodeAccessDenied, value, causes...)
}

func NewInsufficientAuthError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorSubTypeCodeInsufficientAuth, value, causes...)
}

func NewMissingCsrfTokenError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorCodeMissingCsrfToken, value, causes...)
}

func NewInvalidCsrfTokenError(value interface{}, causes ...interface{}) error {
	return NewCodedError(ErrorCodeInvalidCsrfToken, value, causes...)
}
