package opaaccess

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"net/http"
	"time"
)

var logger = log.New("OPA.Access")

// DecisionMakerWithOPA is an access.DecisionMakerFunc that utilize OPA engine
func DecisionMakerWithOPA(opaEngine *sdk.OPA, policy string) access.DecisionMakerFunc {
	return func(ctx context.Context, req *http.Request) (handled bool, decision error) {
		e := AllowRequest(ctx, req, func(opt *RequestQueryOption) {
			opt.OPA = opaEngine
			opt.Policy = policy
		})
		if e != nil {
			return true, security.NewAccessDeniedError(e)
		}
		return true, nil
	}
}

type RequestQueryOptions func(opt *RequestQueryOption)
type RequestQueryOption struct {
	OPA    *sdk.OPA
	Policy string
}

func DefaultRequestQueryOptions() RequestQueryOption {
	return RequestQueryOption{
		OPA:    opa.EmbeddedOPA(),
	}
}

func AllowRequest(ctx context.Context, req *http.Request, opts ...RequestQueryOptions) error {
	opt := DefaultRequestQueryOptions()
	for _, fn := range opts {
		fn(&opt)
	}
	opaOpts := PrepareDecisionQuery(ctx, opt.Policy, req)
	result, e := opt.OPA.Decision(ctx, *opaOpts)
	if e != nil {
		return security.NewAccessDeniedError(e)
	}
	logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, result.Result)
	switch v := result.Result.(type) {
	case bool:
		if v {
			return nil
		}
		return security.NewAccessDeniedError("Access Denied")
	default:
		return security.NewAccessDeniedError(fmt.Errorf("unknow OPA result type %T", result.Result))
	}
}

func PrepareDecisionQuery(ctx context.Context, policy string, req *http.Request) *sdk.DecisionOptions {
	opts := sdk.DecisionOptions{
		Now:                 time.Now(),
		Path:                policy,
		Input:               opa.InputApiAccess{
			Authentication: opa.NewAuthenticationClause(ctx),
			Request:        opa.NewRequestClause(req),
		},
		StrictBuiltinErrors: false,
	}

	if data, e := json.Marshal(opts.Input); e != nil {
		logger.WithContext(ctx).Errorf("Input marshalling error: %v", e)
	} else {
		logger.WithContext(ctx).Debugf("Input: %s", data)
	}
	return &opts
}
