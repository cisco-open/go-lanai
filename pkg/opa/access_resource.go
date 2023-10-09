package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"time"
)

type ResourceQueryOptions func(res *ResourceQuery)

type ResourceQuery struct {
	OPA    *sdk.OPA
	Policy string
	ResourceValues
	Delta    *ResourceValues
	InputCustomizers []InputCustomizer
	// RawInput overrides any input related options
	RawInput interface{}
	// LogLevel override decision log level when presented
	LogLevel *log.LoggingLevel
}

func SilentResourceQuery() ResourceQueryOptions {
	var silent = log.LevelOff
	return func(opt *ResourceQuery) {
		opt.LogLevel = &silent
	}
}

func AllowResource(ctx context.Context, resType string, op ResourceOperation, opts ...ResourceQueryOptions) error {
	res := ResourceQuery{
		OPA: EmbeddedOPA(),
		InputCustomizers: embeddedOPA.inputCustomizers,
		ResourceValues: ResourceValues{ExtraData: map[string]interface{}{}},
	}
	for _, fn := range opts {
		fn(&res)
	}
	if len(res.Policy) == 0 {
		res.Policy = fmt.Sprintf("%s/allow_%v", resType, op)
	}
	ctx = contextWithOverriddenLogLevel(ctx, res.LogLevel)
	opaOpts, e := PrepareResourceDecisionQuery(ctx, res.Policy, resType, op, &res)
	if e != nil {
		return ErrInternal.WithMessage(`error when preparing OPA input: %v`, e)
	}
	result, e := res.OPA.Decision(ctx, *opaOpts)
	return handleDecisionResult(ctx, result, e, "ResourceQuery")
}

func PrepareResourceDecisionQuery(ctx context.Context, policy string, resType string, op ResourceOperation, res *ResourceQuery) (*sdk.DecisionOptions, error) {
	input, e := constructResourceDecisionInput(ctx, resType, op, res)
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

func constructResourceDecisionInput(ctx context.Context, resType string, op ResourceOperation, res *ResourceQuery) (interface{}, error) {
	if res.RawInput != nil {
		return res.RawInput, nil
	}
	input := NewInput()
	input.Authentication = NewAuthenticationClause()
	input.Resource = NewResourceClause(resType, op)
	input.Resource.CurrentResourceValues = CurrentResourceValues(res.ResourceValues)
	input.Resource.Delta = res.Delta

	for _, customizer := range res.InputCustomizers {
		if e := customizer.Customize(ctx, input); e != nil {
			return nil, e
		}
	}
	return input, nil
}
