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
	OPA         *sdk.OPA
	Policy      string
	Unknowns    []string               `json:"-"`
	QueryMapper sdk.PartialQueryMapper `json:"-"`
}

func FilterResource(ctx context.Context, resType string, op ResourceOperation, opts ...ResourceFilterOptions) (*sdk.PartialResult, error) {
	res := ResourceFilter{OPA: EmbeddedOPA()}
	for _, fn := range opts {
		fn(&res)
	}
	if len(res.Policy) == 0 {
		res.Policy = fmt.Sprintf("data.%s.filter_%v", resType, op)
	}
	opaOpts := PrepareResourcePartialQuery(ctx, res.Policy, resType, op, &res)
	result, e := res.OPA.Partial(ctx, *opaOpts)
	if e != nil {
		switch {
		case errors.Is(e, QueriesNotResolvedError):
			return nil, e
		default:
			return nil, InternalError.WithMessage("failed to perform partial evaluation: %v", e)
		}
	}
	logger.WithContext(ctx).Infof("Partial Result [%s]: %v", result.ID, result.AST)
	return result, nil
}

func PrepareResourcePartialQuery(ctx context.Context, policy string, resType string, op ResourceOperation, res *ResourceFilter) *sdk.PartialOptions {
	input := InputApiAccess{
		Authentication: NewAuthenticationClause(ctx),
		Resource:       NewResourceClause(resType, op),
	}

	mapper := res.QueryMapper
	if v, ok := res.QueryMapper.(ContextAwarePartialQueryMapper); ok {
		mapper = v.WithContext(ctx)
	}
	opts := sdk.PartialOptions{
		Now:      time.Now(),
		Input:    &input,
		Query:    policy,
		Unknowns: res.Unknowns,
		Mapper:   mapper,
	}

	if data, e := json.Marshal(opts.Input); e != nil {
		logger.WithContext(ctx).Errorf("InputField marshalling error: %v", e)
	} else {
		logger.WithContext(ctx).Debugf("InputField: %s", data)
	}
	return &opts
}
