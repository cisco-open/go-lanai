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
	ExtraData   map[string]interface{}
}

func AllowRequest(ctx context.Context, req *http.Request, opts ...RequestQueryOptions) error {
	opt := RequestQueryOption{OPA: EmbeddedOPA(), ExtraData: map[string]interface{}{}}
	for _, fn := range opts {
		fn(&opt)
	}
	opaOpts := PrepareRequestDecisionQuery(ctx, opt.Policy, req, opt.ExtraData)
	result, e := opt.OPA.Decision(ctx, *opaOpts)
	switch {
	case sdk.IsUndefinedErr(e):
		logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, "not true")
		return ErrAccessDenied
	case e != nil:
		return ErrAccessDenied.WithMessage("unable to execute OPA query: %v", e)
	}
	logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, result.Result)
	switch v := result.Result.(type) {
	case bool:
		if v {
			return nil
		}
		return ErrAccessDenied
	default:
		return ErrAccessDenied.WithMessage("unsupported OPA result type %T", result.Result)
	}
}

func PrepareRequestDecisionQuery(ctx context.Context, policy string, req *http.Request, extra map[string]interface{}) *sdk.DecisionOptions {
	input := NewInput()
	input.Authentication = NewAuthenticationClause(ctx)
	input.Request = NewRequestClause(req)
	input.Request.ExtraData = extra
	opts := sdk.DecisionOptions{
		Now:                 time.Now(),
		Path:                policy,
		Input:               input,
		StrictBuiltinErrors: false,
	}

	if data, e := json.Marshal(opts.Input); e != nil {
		logger.WithContext(ctx).Errorf("Input marshalling error: %v", e)
	} else {
		logger.WithContext(ctx).Debugf("Input: %s", data)
	}
	return &opts
}

