package opa

import (
	"context"
	"github.com/open-policy-agent/opa/sdk"
)

func handleDecisionResult(ctx context.Context, result *sdk.DecisionResult, err error, targetName string) error {
	switch {
	case sdk.IsUndefinedErr(err):
		logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, "not true")
		return errorWithTargetName(targetName)
	case err != nil:
		return ErrAccessDenied.WithMessage("unable to execute OPA query: %v", err)
	}

	switch v := result.Result.(type) {
	case bool:
		if !v {
			logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, result.Result)
			return errorWithTargetName(targetName)
		}
	default:
		logger.WithContext(ctx).Infof("Decision [%s]: %v", result.ID, result.Result)
		return ErrAccessDenied.WithMessage("unsupported OPA result type %T", result.Result)
	}
	return nil
}

func errorWithTargetName(targetName string) error {
	if len(targetName) == 0 {
		return ErrAccessDenied
	}
	return ErrAccessDenied.WithMessage("%s Access Denied", targetName)
}
