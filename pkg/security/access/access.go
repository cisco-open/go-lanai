package access

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"net/http"
)

// DecisionMakerFunc determine if current user can access to given http.Request
// if the given request is not handled by this function, return false, nil
// if the given request is handled and the access is granted, return true, nil
// otherwise, return true, security.ErrorTypeCodeAccessControl error
type DecisionMakerFunc func(context.Context, *http.Request) (handled bool, decision error)

// AcrMatcher short for Access Control Request Matcher, accepts *http.Request or http.Request
type AcrMatcher matcher.RequestMatcher

// ControlFunc make access control decision based on security.Authentication
type ControlFunc func(security.Authentication) bool

func MakeDecisionMakerFunc(matcher AcrMatcher, cf ControlFunc) DecisionMakerFunc {
	return func(ctx context.Context, r *http.Request) (bool, error) {
		matches, err := matcher.Matches(r)
		if !matches || err != nil {
			return false, err
		}

		auth := security.Get(ctx)
		if cf(auth) {
			return true, nil
		} else {
			return true, security.NewAccessControlError("access denied")
		}
	}
}