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
	"testing"
)

func TestTypeComparison(t *testing.T) {
	switch {
	case !errors.Is(ErrorSubTypeInternalError, ErrorTypeSecurity):
		t.Errorf("ErrorType should match ErrorTypeSecurity")

	case !errors.Is(ErrorSubTypeInternalError, ErrorSubTypeInternalError):
		t.Errorf("ErrorType should match itself")

	case !errors.Is(ErrorSubTypeInternalError, ErrorTypeAuthentication):
		t.Errorf("ErrorSubType should match its own ErrorType")

	case errors.Is(ErrorTypeAuthentication, ErrorSubTypeInternalError):
		t.Errorf("ErrorType should not match its own ErrorSubType")

	case errors.Is(ErrorTypeAuthentication, ErrorTypeAccessControl):
		t.Errorf("Different ErrorType should not match each other")

	case errors.Is(ErrorSubTypeInternalError, ErrorSubTypeUsernamePasswordAuth):
		t.Errorf("Different ErrorSubType should not match each other")

	case !errors.Is(ErrorSubTypeCsrf, ErrorTypeAccessControl):
		t.Errorf("ErrorSubTypeCsrf should be ErrorTypeAccessControl error")

	case !errors.Is(ErrorTypeInternal, ErrorTypeSecurity):
		t.Errorf("ErrorTypeInternal should be ErrorTypeSecurity")
	}
}

func TestAuthenticatorNotAvailableError(t *testing.T) {
	ana := NewAuthenticatorNotAvailableError("not available")
	another := NewAuthenticatorNotAvailableError("different message")
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(ana, ErrorTypeSecurity):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorTypeSecurity")

	case !errors.Is(ana, ErrorTypeAuthentication):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorTypeAuthentication")

	case !errors.Is(ana, ErrorSubTypeInternalError):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorSubTypeInternalError")

	case !errors.Is(ana, another):
		t.Errorf("Two NewAuthenticatorNotAvailableError should match each other")

	case errors.Is(ana, nonCoded):
		t.Errorf("NewAuthenticatorNotAvailableError should not match non-coded error")

	case errors.Is(ana, ErrorTypeAccessControl):
		t.Errorf("NewAuthenticatorNotAvailableError should not match ErrorTypeAccessControl")

	case errors.Is(ana, ErrorSubTypeUsernamePasswordAuth):
		t.Errorf("NewAuthenticatorNotAvailableError should not match ErrorSubTypeUsernamePasswordAuth")
	}
}

func TestBadCredentialsError(t *testing.T) {
	coded := NewBadCredentialsError("wrong password")
	another := NewBadCredentialsError("different message")
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(coded, ErrorTypeSecurity):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorTypeSecurity")

	case !errors.Is(coded, ErrorTypeAuthentication):
		t.Errorf("NewBadCredentialsError should match ErrorTypeAuthentication")

	case !errors.Is(coded, ErrorSubTypeUsernamePasswordAuth):
		t.Errorf("NewBadCredentialsError should match ErrorSubTypeUsernamePasswordAuth")

	case !errors.Is(coded, another):
		t.Errorf("Two NewBadCredentialsError should match each other")

	case errors.Is(coded, nonCoded):
		t.Errorf("NewBadCredentialsError should not match non-coded error")

	case errors.Is(coded, ErrorTypeAccessControl):
		t.Errorf("NewBadCredentialsError should not match ErrorTypeAccessControl")

	case errors.Is(coded, ErrorSubTypeInternalError):
		t.Errorf("NewBadCredentialsError should not match ErrorSubTypeInternalError")
	}
}

func TestMissingCsrfTokenError(t *testing.T) {
	coded := NewMissingCsrfTokenError("missing csrf token")
	another := NewMissingCsrfTokenError("different message")
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(coded, ErrorTypeSecurity):
		t.Errorf("NewMissingCsrfTokenError should match ErrorTypeSecurity")

	case !errors.Is(coded, ErrorTypeAccessControl):
		t.Errorf("NewMissingCsrfTokenError should match ErrorTypeAccessControl")

	case !errors.Is(coded, ErrorSubTypeCsrf):
		t.Errorf("NewMissingCsrfTokenError should match ErrorSubTypeCsrf")

	case !errors.Is(coded, another):
		t.Errorf("Two NewMissingCsrfTokenError should match each other")

	case errors.Is(coded, nonCoded):
		t.Errorf("NewMissingCsrfTokenError should not match non-coded error")

	case errors.Is(coded, ErrorTypeAuthentication):
		t.Errorf("NewMissingCsrfTokenError should not match ErrorTypeAuthentication")
	}
}

func TestInvalidCsrfTokenError(t *testing.T) {
	coded := NewInvalidCsrfTokenError("invalid csrf token")
	another := NewInvalidCsrfTokenError("different message")
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(coded, ErrorTypeSecurity):
		t.Errorf("NewInvalidCsrfTokenError should match ErrorTypeSecurity")

	case !errors.Is(coded, ErrorTypeAccessControl):
		t.Errorf("NewInvalidCsrfTokenError should match ErrorTypeAccessControl")

	case !errors.Is(coded, ErrorSubTypeCsrf):
		t.Errorf("NewInvalidCsrfTokenError should match ErrorSubTypeCsrf")

	case !errors.Is(coded, another):
		t.Errorf("Two NewInvalidCsrfTokenError should match each other")

	case errors.Is(coded, nonCoded):
		t.Errorf("NewInvalidCsrfTokenError should not match non-coded error")

	case errors.Is(coded, ErrorTypeAuthentication):
		t.Errorf("NewInvalidCsrfTokenError should not match ErrorTypeAuthentication")
	}
}

func TestInternalError(t *testing.T) {
	coded := NewInternalError("some internal error")
	switch {
	case !errors.Is(coded, ErrorTypeInternal):
		t.Errorf("NewInternalError should be ErrorTypeInternal")
	case !errors.Is(coded, ErrorTypeSecurity):
		t.Errorf("NewInternalError should be ErrorTypeSecurity")
	case errors.Is(coded, ErrorTypeAuthentication):
		t.Errorf("NewInternalError should not be ErrorTypeAuthentication")
	}
}