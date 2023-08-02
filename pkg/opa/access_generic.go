package opa

import (
	"context"
	"encoding/json"
	"github.com/open-policy-agent/opa/sdk"
	"time"
)

type QueryOptions func(q *Query)

type Query struct {
	OPA              *sdk.OPA
	Policy           string
	InputCustomizers []InputCustomizer
	RawInput         interface{}
}

func QueryWithPolicy(policy string) QueryOptions {
	return func(q *Query) {
		q.Policy = policy
	}
}

func QueryWithInputCustomizer(customizer InputCustomizerFunc) QueryOptions {
	return func(q *Query) {
		q.InputCustomizers = append(q.InputCustomizers, customizer)
	}
}

// Allow is generic API for querying policy. This function only populate minimum input data like authentication.
// For more specialized function, see AllowResource, AllowRequest, etc.
func Allow(ctx context.Context, opts ...QueryOptions) error {
	query := Query{
		OPA:              EmbeddedOPA(),
		InputCustomizers: embeddedOPA.inputCustomizers,
	}
	for _, fn := range opts {
		fn(&query)
	}
	if len(query.Policy) == 0 {
		return ErrInternal.WithMessage("policy is required for generic Allow function")
	}
	opaOpts, e := PrepareGenericDecisionQuery(ctx, &query)
	if e != nil {
		return ErrInternal.WithMessage(`error when preparing OPA input: %v`, e)
	}
	result, e := query.OPA.Decision(ctx, *opaOpts)
	return handleDecisionResult(ctx, result, e, "")
}

func PrepareGenericDecisionQuery(ctx context.Context, query *Query) (*sdk.DecisionOptions, error) {
	input, e := constructGenericDecisionInput(ctx, query)
	if e != nil {
		return nil, e
	}
	opts := sdk.DecisionOptions{
		Now:                 time.Now(),
		Path:                query.Policy,
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

func constructGenericDecisionInput(ctx context.Context, query *Query) (interface{}, error) {
	if query.RawInput != nil {
		return query.RawInput, nil
	}
	input := NewInput()
	input.Authentication = NewAuthenticationClause()
	for _, customizer := range query.InputCustomizers {
		if e := customizer.Customize(ctx, input); e != nil {
			return nil, e
		}
	}
	return input, nil
}
