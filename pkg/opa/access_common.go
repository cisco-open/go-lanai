// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/open-policy-agent/opa/sdk"
)

func contextWithOverriddenLogLevel(ctx context.Context, override *log.LoggingLevel) context.Context {
	if override == nil {
		return ctx
	}
	return logContextWithLevel(ctx, *override)
}

func handleDecisionResult(ctx context.Context, result *sdk.DecisionResult, rErr error, targetName string) (err error) {
	var parsedResult interface{}
	defer func() {
		event := &resultEvent{
			Result: parsedResult,
			Deny:   err != nil,
		}
		if result != nil {
			event.ID = result.ID
		}
		if err == nil {
			eventLogger(ctx, log.LevelDebug).WithKV(kLogDecisionReason, event).Printf("Allow [%v]", event.ID)
		} else {
			eventLogger(ctx, log.LevelDebug).WithKV(kLogDecisionReason, event).Printf("Deny [%v]", event.ID)
		}
	}()

	switch {
	case sdk.IsUndefinedErr(rErr):
		parsedResult = "not true"
		return errorWithTargetName(targetName)
	case rErr != nil:
		parsedResult = rErr
		return ErrAccessDenied.WithMessage("unable to execute OPA query: %v", rErr)
	}

	parsedResult = result.Result
	switch v := result.Result.(type) {
	case bool:
		if !v {
			return errorWithTargetName(targetName)
		}
	default:
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
