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
// "decision" indicate whether the access is grated
// "reason" is optional and is used when access is denied. if not specified, security.NewAccessDeniedError will be used
type ControlFunc func(security.Authentication) (decision bool, reason error)

func MakeDecisionMakerFunc(matcher AcrMatcher, cf ControlFunc) DecisionMakerFunc {
	return func(ctx context.Context, r *http.Request) (bool, error) {
		matches, err := matcher.Matches(r)
		if !matches || err != nil {
			return false, err
		}

		auth := security.Get(ctx)
		granted, reason := cf(auth)
		if granted {
			return true, nil
		} else if reason != nil {
			return true, reason
		} else {
			return true, security.NewAccessDeniedError("access denied")
		}
	}
}