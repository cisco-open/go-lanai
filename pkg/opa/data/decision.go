package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"time"
)

var logger = log.New("OPA.Data")

type ResourceOptions func(res *Resource)

type Resource struct {
	OPA         *sdk.OPA
	TenantID    string                 `json:"tenant_id,omitempty"`
	TenantPath  []string               `json:"tenant_path,omitempty"`
	OwnerID     string                 `json:"owner_id,omitempty"`
	Share       map[string][]string    `json:"share,omitempty"`
}

func AllowResource(ctx context.Context, resType string, op opa.ResourceOperation, opts ...ResourceOptions) error {
	res := Resource{OPA: opa.EmbeddedOPA()}
	for _, fn := range opts {
		fn(&res)
	}
	policy := fmt.Sprintf("%s/allow_%v", resType, op)
	opaOpts := PrepareDecisionQuery(ctx, policy, resType, op, &res)
	result, e := res.OPA.Decision(ctx, *opaOpts)
	if e != nil {
		return security.NewAccessDeniedError(e)
	}
	logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, result.Result)
	switch v := result.Result.(type) {
	case bool:
		if !v {
			return security.NewAccessDeniedError("Resource Denied")
		}
	default:
		return security.NewAccessDeniedError(fmt.Errorf("unknow OPA result type %T", result.Result))
	}
	return nil
}

func PrepareDecisionQuery(ctx context.Context, policy string, resType string, op opa.ResourceOperation, res *Resource) *sdk.DecisionOptions {
	input := opa.InputApiAccess{
		Authentication: opa.NewAuthenticationClause(ctx),
		Resource:       opa.NewResourceClause(resType, op),
	}
	input.Resource.TenantID = res.TenantID
	input.Resource.OwnerID = res.OwnerID
	input.Resource.TenantPath = res.TenantPath
	input.Resource.Share = res.Share

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
