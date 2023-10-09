package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
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
	// LogLevel override decision log level when presented
	LogLevel *log.LoggingLevel
}

func RequestQueryWithPolicy(policy string) RequestQueryOptions {
	return func(opt *RequestQuery) {
		opt.Policy = policy
	}
}

func SilentRequestQuery() RequestQueryOptions {
	var silent = log.LevelOff
	return func(opt *RequestQuery) {
		opt.LogLevel = &silent
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
	ctx = contextWithOverriddenLogLevel(ctx, opt.LogLevel)
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

	//if data, e := json.Marshal(opts.Input); e != nil {
	//	eventLogger(ctx, log.LevelError).Printf("Input marshalling error: %v", e)
	//} else {
	//	eventLogger(ctx, log.LevelDebug).Printf("Input: %s", data)
	//}
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
