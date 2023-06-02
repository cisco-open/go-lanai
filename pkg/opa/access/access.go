package opaaccess

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"encoding/json"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"net/http"
	"time"
)

var logger = log.New("OPA.Access")

// DecisionMakerWithOPA is an access.DecisionMakerFunc that utilize OPA engine
func DecisionMakerWithOPA(opa *sdk.OPA, policy string) access.DecisionMakerFunc {
	return func(ctx context.Context, req *http.Request) (handled bool, decision error) {
		opts := PrepareOpaQuery(ctx, policy, req)
		result, e := opa.Decision(ctx, *opts)
		if e != nil {
			return true, security.NewAccessDeniedError(e)
		}
		logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, result.Result)
		switch v := result.Result.(type) {
		case bool:
			if v {
				return true, nil
			}
			return true, security.NewAccessDeniedError("Access Denied")
		default:
			return true, security.NewAccessDeniedError(fmt.Errorf("unknow OPA result type %T", result.Result))
		}
	}
}

func PrepareOpaQuery(ctx context.Context, policy string, req *http.Request) *sdk.DecisionOptions {
	opts := sdk.DecisionOptions{
		Now:                 time.Now(),
		Path:                policy,
		Input:               opa.InputApiAccess{
			Authentication: opa.NewAuthenticationClause(ctx),
			Request:        opa.NewRequestClause(req),
		},
		StrictBuiltinErrors: false,
	}

	if data, e := json.Marshal(opts.Input); e != nil {
		logger.WithContext(ctx).Errorf("Input marshalling error: %v", e)
	} else {
		logger.WithContext(ctx).Debugf("Input: %s", data)
	}
	return &opts
}
