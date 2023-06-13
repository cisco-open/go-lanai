package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"time"
)

type ResourceFilterOptions func(rf *ResourceFilter)

type ResourceFilter struct {
	OPA         *sdk.OPA
	TenantID    string                         `json:"tenant_id,omitempty"`
	TenantPath  []string                       `json:"tenant_path,omitempty"`
	OwnerID     string                         `json:"owner_id,omitempty"`
	Share       map[string][]string            `json:"share,omitempty"`
	Unknowns    []string                       `json:"-"`
	QueryMapper ContextAwarePartialQueryMapper `json:"-"`
}

func FilterResource(ctx context.Context, resType string, op opa.ResourceOperation, opts ...ResourceFilterOptions) (*sdk.PartialResult, error) {
	res := ResourceFilter{OPA: opa.EmbeddedOPA()}
	for _, fn := range opts {
		fn(&res)
	}
	policy := fmt.Sprintf("data.%s.filter_%v = true", resType, op)
	opaOpts := PreparePartialQuery(ctx, policy, resType, op, &res)
	result, e := res.OPA.Partial(ctx, *opaOpts)
	if e != nil {
		return nil, fmt.Errorf("failed to perform partial evaluation: %v", e)
	}
	logger.WithContext(ctx).Infof("Partial Result [%s]: %v", result.ID, result.AST)
	return result, nil
}

func PreparePartialQuery(ctx context.Context, policy string, resType string, op opa.ResourceOperation, res *ResourceFilter) *sdk.PartialOptions {
	input := opa.InputApiAccess{
		Authentication: opa.NewAuthenticationClause(ctx),
		Resource:       opa.NewResourceClause(resType, op),
	}
	input.Resource.TenantID = res.TenantID
	input.Resource.OwnerID = res.OwnerID
	input.Resource.TenantPath = res.TenantPath
	input.Resource.Share = res.Share

	opts := sdk.PartialOptions{
		Now:   time.Now(),
		Input: &input,
		Query: policy,
		Unknowns: res.Unknowns,
		Mapper: res.QueryMapper.WithContext(ctx),
	}

	if data, e := json.Marshal(opts.Input); e != nil {
		logger.WithContext(ctx).Errorf("Input marshalling error: %v", e)
	} else {
		logger.WithContext(ctx).Debugf("Input: %s", data)
	}
	return &opts
}
