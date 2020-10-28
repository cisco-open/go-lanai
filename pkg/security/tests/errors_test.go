package tests

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"errors"
	"testing"
)

func TestTypeComparison(t *testing.T) {
	switch {
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
	}
}

func TestAuthenticatorNotAvailableError(t *testing.T) {
	ana := security.NewAuthenticatorNotAvailableError("not available")
	another := security.NewAuthenticatorNotAvailableError("different message")
	nonCoded := errors.New("non-coded error")
	switch {
	case !errors.Is(ana, security.ErrorTypeAuthentication):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorTypeAuthentication")

	case !errors.Is(ana, security.ErrorSubTypeInternalError):
		t.Errorf("NewAuthenticatorNotAvailableError should match ErrorSubTypeInternalError")

	case !errors.Is(ana, another):
		t.Errorf("Two NewAuthenticatorNotAvailableError should match each other")

	case errors.Is(ana, nonCoded):
		t.Errorf("NewAuthenticatorNotAvailableError should not match non-coded errors")

	case errors.Is(ana, security.ErrorTypeAccessControl):
		t.Errorf("NewAuthenticatorNotAvailableError should not match ErrorTypeAccessControl")

	case errors.Is(ana, security.ErrorSubTypeUsernamePasswordAuth):
		t.Errorf("NewAuthenticatorNotAvailableError should not match ErrorSubTypeUsernamePasswordAuth")
	}
}
