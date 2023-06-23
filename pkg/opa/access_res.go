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
	OPA        *sdk.OPA
	ResourceValues
	Delta *ResourceValues
}

func AllowResource(ctx context.Context, resType string, op ResourceOperation, opts ...ResourceOptions) error {
	res := Resource{OPA: EmbeddedOPA(),
		ResourceValues: ResourceValues{ExtraData: map[string]interface{}{}},
	}
	for _, fn := range opts {
		fn(&res)
	}
	policy := fmt.Sprintf("%s/allow_%v", resType, op)
	opaOpts := PrepareResourceDecisionQuery(ctx, policy, resType, op, &res)
	result, e := res.OPA.Decision(ctx, *opaOpts)
	if e != nil {
		return ErrAccessDenied.WithMessage("unable to execute OPA query: %v", e)
	}
	logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, result.Result)
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
	input := NewInput()
	input.Authentication = NewAuthenticationClause(ctx)
	input.Resource = NewResourceClause(resType, op)

	input.Resource.CurrentResourceValues = CurrentResourceValues(res.ResourceValues)
	input.Resource.Delta = res.Delta

	opts := sdk.DecisionOptions{
		Now:                 time.Now(),
		Path:                policy,
		Input:               &input,
		StrictBuiltinErrors: false,
	}

	if data, e := json.Marshal(opts.Input); e != nil {
		logger.WithContext(ctx).Errorf("Input marshalling error: %v", e)
	} else {
		logger.WithContext(ctx).Debugf("Input: %s", data)
	}
	return &opts
}
