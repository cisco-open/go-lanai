package opa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"time"
)

type ContextAwarePartialQueryMapper interface {
	sdk.PartialQueryMapper
	WithContext(ctx context.Context) sdk.PartialQueryMapper
	Context() context.Context
}

type ResourceFilterOptions func(rf *ResourceFilter)

type ResourceFilter struct {
	OPA      *sdk.OPA
	Query    string
	Unknowns []string
	QueryMapper      sdk.PartialQueryMapper
	Delta            *ResourceValues
	ExtraData        map[string]interface{}
	InputCustomizers []InputCustomizer
	// RawInput overrides any input related options
	RawInput interface{}
}

func FilterResource(ctx context.Context, resType string, op ResourceOperation, opts ...ResourceFilterOptions) (*sdk.PartialResult, error) {
	res := ResourceFilter{
		OPA:              EmbeddedOPA(),
		InputCustomizers: embeddedOPA.inputCustomizers,
		QueryMapper:      &sdk.RawMapper{},
		ExtraData:        map[string]interface{}{},
	}
	for _, fn := range opts {
		fn(&res)
	}
	if len(res.Query) == 0 {
		res.Query = fmt.Sprintf("data.%s.filter_%v", resType, op)
	}
	opaOpts, e := PrepareResourcePartialQuery(ctx, res.Query, resType, op, &res)
	if e != nil {
		return nil, ErrInternal.WithMessage(`error when preparing OPA input: %v`, e)
	}
	result, e := res.OPA.Partial(ctx, *opaOpts)
	if e != nil {
		switch {
		case sdk.IsUndefinedErr(e):
			return nil, ErrAccessDenied
		case errors.Is(e, ErrQueriesNotResolved):
			return nil, ErrAccessDenied.WithMessage(e.Error())
		default:
			return nil, ErrAccessDenied.WithMessage("failed to perform partial evaluation: %v", e)
		}
	}
	logger.WithContext(ctx).Infof("Partial Result [%s]: %v", result.ID, result.AST)
	return result, nil
}

func PrepareResourcePartialQuery(ctx context.Context, policy string, resType string, op ResourceOperation, res *ResourceFilter) (*sdk.PartialOptions, error) {
	input, e := constructResourcePartialInput(ctx, resType, op, res)
	if e != nil {
		return nil, e
	}
	mapper := res.QueryMapper
	if v, ok := res.QueryMapper.(ContextAwarePartialQueryMapper); ok {
		mapper = v.WithContext(ctx)
	}
	opts := sdk.PartialOptions{
		Now:      time.Now(),
		Input:    input,
		Query:    policy,
		Unknowns: res.Unknowns,
		Mapper:   mapper,
	}

	if data, e := json.Marshal(opts.Input); e != nil {
		logger.WithContext(ctx).Errorf("Input marshalling error: %v", e)
	} else {
		logger.WithContext(ctx).Debugf("Input: %s", data)
	}
	return &opts, nil
}

func constructResourcePartialInput(ctx context.Context, resType string, op ResourceOperation, res *ResourceFilter) (interface{}, error) {
	if res.RawInput != nil {
		return res.RawInput, nil
	}
	input := NewInput()
	input.Authentication = NewAuthenticationClause()
	input.Resource = NewResourceClause(resType, op)
	input.Resource.ExtraData = res.ExtraData
	input.Resource.Delta = res.Delta

	for _, customizer := range res.InputCustomizers {
		if e := customizer.Customize(ctx, input); e != nil {
			return nil, e
		}
	}
	return input, nil
}



