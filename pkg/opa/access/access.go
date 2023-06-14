package opaaccess

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"github.com/open-policy-agent/opa/sdk"
	"net/http"
)

// DecisionMakerWithOPA is an access.DecisionMakerFunc that utilize OPA engine
func DecisionMakerWithOPA(opaEngine *sdk.OPA, policy string) access.DecisionMakerFunc {
	return func(ctx context.Context, req *http.Request) (handled bool, decision error) {
		e := opa.AllowRequest(ctx, req, func(opt *opa.RequestQueryOption) {
			opt.OPA = opaEngine
			opt.Policy = policy
		})
		if e != nil {
			return true, security.NewAccessDeniedError(e)
		}
		return true, nil
	}
}

