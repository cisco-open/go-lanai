package opa

import (
	"context"
	"encoding/json"
	"github.com/open-policy-agent/opa/sdk"
	"net/http"
	"time"
)

type RequestQueryOptions func(opt *RequestQueryOption)
type RequestQueryOption struct {
	OPA    *sdk.OPA
	Policy string
}

func AllowRequest(ctx context.Context, req *http.Request, opts ...RequestQueryOptions) error {
	opt := RequestQueryOption{OPA: EmbeddedOPA()}
	for _, fn := range opts {
		fn(&opt)
	}
	opaOpts := PrepareRequestDecisionQuery(ctx, opt.Policy, req)
	result, e := opt.OPA.Decision(ctx, *opaOpts)
	if e != nil {
		return AccessDeniedError.WithMessage("unable to execute OPA query: %v", e)
	}
	logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, result.Result)
	switch v := result.Result.(type) {
	case bool:
		if v {
			return nil
		}
		return AccessDeniedError
	default:
		return AccessDeniedError.WithMessage("unsupported OPA result type %T", result.Result)
	}
}

func PrepareRequestDecisionQuery(ctx context.Context, policy string, req *http.Request) *sdk.DecisionOptions {
	opts := sdk.DecisionOptions{
		Now:                 time.Now(),
		Path:                policy,
		Input:               InputApiAccess{
			Authentication: NewAuthenticationClause(ctx),
			Request:        NewRequestClause(req),
		},
		StrictBuiltinErrors: false,
	}

	if data, e := json.Marshal(opts.Input); e != nil {
		logger.WithContext(ctx).Errorf("InputField marshalling error: %v", e)
	} else {
		logger.WithContext(ctx).Debugf("InputField: %s", data)
	}
	return &opts
}

