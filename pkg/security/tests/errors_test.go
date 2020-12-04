package tests

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"errors"
	"testing"
)

func TestTypeComparison(t *testing.T) {
	switch {
	case !errors.Is(security.ErrorSubTypeInternalError, security.ErrorTypeSecurity):
		t.Errorf("ErrorType should match ErrorTypeSecurity")

	case !errors.Is(security.ErrorSubTypeInternalError, security.ErrorSubTypeInternalError):
		t.Errorf("ErrorType should match itself")

	case !errors.Is(security.ErrorSubTypeInternalError, security.ErrorTypeAuthentication):
		t.Errorf("ErrorSubType should match its own ErrorType")

	case errors.Is(security.ErrorTypeAuthentication, security.ErrorSubTypeInternalError):
		t.Errorf("ErrorType should not match its own ErrorSubType")

	case errors.Is(security.ErrorTypeAuthentication, security.ErrorTypeAccessControl):
		t.Errorf("Different ErrorType should not match each other")

	case errors.Is(security.ErrorSubTypeInternalError, security.ErrorSubTypeUsernamePasswordAuth):
		t.Errorf("Different ErrorSubType should not match each other")

	case !errors.Is(security.ErrorSubTypeCsrf, security.ErrorTypeAccessControl):
		t.Errorf("ErrorSubTypeCsrf should be ErrorTypeAccessControl error")
	}
}

func TestAuthenticatorNotAvailableError(t *testing.T) {
	ana := security.NewAuthenticatorNotAvailableError("not available")
	another := security.NewAuthenticatorNotAvailableError("different message")
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(ana, security.ErrorTypeSecurity):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorTypeSecurity")

	case !errors.Is(ana, security.ErrorTypeAuthentication):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorTypeAuthentication")

	case !errors.Is(ana, security.ErrorSubTypeInternalError):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorSubTypeInternalError")

	case !errors.Is(ana, another):
		t.Errorf("Two NewAuthenticatorNotAvailableError should match each other")

	case errors.Is(ana, nonCoded):
		t.Errorf("NewAuthenticatorNotAvailableError should not match non-coded error")

	case errors.Is(ana, security.ErrorTypeAccessControl):
		t.Errorf("NewAuthenticatorNotAvailableError should not match ErrorTypeAccessControl")

	case errors.Is(ana, security.ErrorSubTypeUsernamePasswordAuth):
		t.Errorf("NewAuthenticatorNotAvailableError should not match ErrorSubTypeUsernamePasswordAuth")
	}
}

func TestBadCredentialsError(t *testing.T) {
	coded := security.NewBadCredentialsError("wrong password")
	another := security.NewBadCredentialsError("different message")
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(coded, security.ErrorTypeSecurity):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorTypeSecurity")

	case !errors.Is(coded, security.ErrorTypeAuthentication):
		t.Errorf("NewBadCredentialsError should match ErrorTypeAuthentication")

	case !errors.Is(coded, security.ErrorSubTypeUsernamePasswordAuth):
		t.Errorf("NewBadCredentialsError should match ErrorSubTypeUsernamePasswordAuth")

	case !errors.Is(coded, another):
		t.Errorf("Two NewBadCredentialsError should match each other")

	case errors.Is(coded, nonCoded):
		t.Errorf("NewBadCredentialsError should not match non-coded error")

	case errors.Is(coded, security.ErrorTypeAccessControl):
		t.Errorf("NewBadCredentialsError should not match ErrorTypeAccessControl")

	case errors.Is(coded, security.ErrorSubTypeInternalError):
		t.Errorf("NewBadCredentialsError should not match ErrorSubTypeInternalError")
	}
}

func TestMissingCsrfTokenError(t *testing.T) {
	coded := security.NewMissingCsrfTokenError("missing csrf token")
	another := security.NewMissingCsrfTokenError("different message")
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(coded, security.ErrorTypeSecurity):
		t.Errorf("NewMissingCsrfTokenError should match ErrorTypeSecurity")

	case !errors.Is(coded, security.ErrorTypeAccessControl):
		t.Errorf("NewMissingCsrfTokenError should match ErrorTypeAccessControl")

	case !errors.Is(coded, security.ErrorSubTypeCsrf):
		t.Errorf("NewMissingCsrfTokenError should match ErrorSubTypeCsrf")

	case !errors.Is(coded, another):
		t.Errorf("Two NewMissingCsrfTokenError should match each other")

	case errors.Is(coded, nonCoded):
		t.Errorf("NewMissingCsrfTokenError should not match non-coded error")

	case errors.Is(coded, security.ErrorTypeAuthentication):
		t.Errorf("NewMissingCsrfTokenError should not match ErrorTypeAuthentication")
	}
}

func TestInvalidCsrfTokenError(t *testing.T) {
	coded := security.NewInvalidCsrfTokenError("invalid csrf token")
	another := security.NewInvalidCsrfTokenError("different message")
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(coded, security.ErrorTypeSecurity):
		t.Errorf("NewInvalidCsrfTokenError should match ErrorTypeSecurity")

	case !errors.Is(coded, security.ErrorTypeAccessControl):
		t.Errorf("NewInvalidCsrfTokenError should match ErrorTypeAccessControl")

	case !errors.Is(coded, security.ErrorSubTypeCsrf):
		t.Errorf("NewInvalidCsrfTokenError should match ErrorSubTypeCsrf")

	case !errors.Is(coded, another):
		t.Errorf("Two NewInvalidCsrfTokenError should match each other")

	case errors.Is(coded, nonCoded):
		t.Errorf("NewInvalidCsrfTokenError should not match non-coded error")

	case errors.Is(coded, security.ErrorTypeAuthentication):
		t.Errorf("NewInvalidCsrfTokenError should not match ErrorTypeAuthentication")
	}
}