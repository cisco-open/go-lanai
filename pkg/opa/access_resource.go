package opa

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"time"
)

type ResourceOptions func(res *Resource)

type Resource struct {
	OPA    *sdk.OPA
	Policy string
	ResourceValues
	Delta    *ResourceValues
	RawInput interface{}
}

func AllowResource(ctx context.Context, resType string, op ResourceOperation, opts ...ResourceOptions) error {
	res := Resource{OPA: EmbeddedOPA(),
		ResourceValues: ResourceValues{ExtraData: map[string]interface{}{}},
	}
	for _, fn := range opts {
		fn(&res)
	}
	if len(res.Policy) == 0 {
		res.Policy = fmt.Sprintf("%s/allow_%v", resType, op)
	}
	opaOpts := PrepareResourceDecisionQuery(ctx, res.Policy, resType, op, &res)
	result, e := res.OPA.Decision(ctx, *opaOpts)
	switch {
	case sdk.IsUndefinedErr(e):
		logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, "not true")
		return ErrAccessDenied.WithMessage("Resource Access Denied")
	case e != nil:
		return ErrAccessDenied.WithMessage("unable to execute OPA query: %v", e)
	}

	switch v := result.Result.(type) {
	case bool:
		if !v {
			return ErrAccessDenied.WithMessage("Resource Access Denied")
		}
	default:
		return ErrAccessDenied.WithMessage("unsupported OPA result type %T", result.Result)
	}
	return nil
}

func PrepareResourceDecisionQuery(ctx context.Context, policy string, resType string, op ResourceOperation, res *Resource) *sdk.DecisionOptions {
	input := constructResourceDecisionInput(ctx, resType, op, res)
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

func constructResourceDecisionInput(ctx context.Context, resType string, op ResourceOperation, res *Resource) interface{} {
	if res.RawInput != nil {
		return res.RawInput
	}
	input := NewInput()
	input.Authentication = NewAuthenticationClause(ctx)
	input.Resource = NewResourceClause(resType, op)

	input.Resource.CurrentResourceValues = CurrentResourceValues(res.ResourceValues)
	input.Resource.Delta = res.Delta
	return input
}
