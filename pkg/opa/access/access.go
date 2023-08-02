package opaaccess

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"net/http"
)

// DecisionMakerWithOPA is an access.DecisionMakerFunc that utilize OPA engine
func DecisionMakerWithOPA(opts ...opa.RequestQueryOptions) access.DecisionMakerFunc {
	return func(ctx context.Context, req *http.Request) (handled bool, decision error) {
		e := opa.AllowRequest(ctx, req, opts...)
		if e != nil {
			return true, security.NewAccessDeniedError(e)
		}
		return true, nil
	}
}


