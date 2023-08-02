package opa

import (
	"context"
	"encoding/json"
	"github.com/open-policy-agent/opa/sdk"
	"net/http"
	"time"
)

type RequestQueryOptions func(opt *RequestQuery)
type RequestQuery struct {
	OPA              *sdk.OPA
	Policy           string
	ExtraData        map[string]interface{}
	InputCustomizers []InputCustomizer
	// RawInput overrides any input related options
	RawInput interface{}
}

func RequestQueryWithPolicy(policy string) RequestQueryOptions {
	return func(opt *RequestQuery) {
		opt.Policy = policy
	}
}

func AllowRequest(ctx context.Context, req *http.Request, opts ...RequestQueryOptions) error {
	opt := RequestQuery{
		OPA:              EmbeddedOPA(),
		InputCustomizers: embeddedOPA.inputCustomizers,
		ExtraData:        map[string]interface{}{},
	}
	for _, fn := range opts {
		fn(&opt)
	}
	opaOpts, e := PrepareRequestDecisionQuery(ctx, opt.Policy, req, &opt)
	if e != nil {
		return ErrInternal.WithMessage(`error when preparing OPA input: %v`, e)
	}
	result, e := opt.OPA.Decision(ctx, *opaOpts)
	return handleDecisionResult(ctx, result, e, "API")
}

func PrepareRequestDecisionQuery(ctx context.Context, policy string, req *http.Request, opt *RequestQuery) (*sdk.DecisionOptions, error) {
	input, e := constructRequestDecisionInput(ctx, req, opt)
	if e != nil {
		return nil, e
	}
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
	return &opts, nil
}

func constructRequestDecisionInput(ctx context.Context, req *http.Request, opt *RequestQuery) (interface{}, error) {
	if opt.RawInput != nil {
		return opt.RawInput, nil
	}
	input := NewInput()
	input.Authentication = NewAuthenticationClause()
	input.Request = NewRequestClause(req)
	input.Request.ExtraData = opt.ExtraData
	for _, customizer := range opt.InputCustomizers {
		if e := customizer.Customize(ctx, input); e != nil {
			return nil, e
		}
	}
	return input, nil
}
