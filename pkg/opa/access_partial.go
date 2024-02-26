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
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/log"
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
	// OPA (Optional) instance to use for evaluation. Default to EmbeddedOPA()
	OPA *sdk.OPA
	// Query (Optional) OPA query to evaluate.
	// Default to `data.resource.<resource_type>.filter_<resource_operation>`
	Query string
	// Unknowns (Required) List of unknown input fields for partial evaluation. Not providing "unknowns" would not
	// result in immediate error, but very like result in access denial.
	Unknowns []string
	// QueryMapper (Optional) Custom sdk.PartialQueryMapper for translating result rego.PartialQueries.
	// By default, partial result is *rego.PartialQueries. QueryMapper can translate it to other structure.
	// e.g. SQL "Where" clause
	QueryMapper sdk.PartialQueryMapper
	// Delta (Optional) Resource's "changed-to" fields and values. Delta is only applicable to "write" operation.
	// OPA policies may have rules on what values the resource's certain fields can be changed to.
	Delta *ResourceValues
	// ExtraData  (Optional) any key-value pairs in ExtraData will be added into query input under `input.resource.*`
	ExtraData map[string]interface{}
	// InputCustomizers customizers to finalize/modify query input before evaluation
	InputCustomizers []InputCustomizer
	// RawInput overrides any input related options
	RawInput interface{}
	// LogLevel override decision log level when presented
	LogLevel *log.LoggingLevel
}

func SilentResourceFilter() ResourceFilterOptions {
	var silent = log.LevelOff
	return func(opt *ResourceFilter) {
		opt.LogLevel = &silent
	}
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
		res.Query = fmt.Sprintf("data.%s.%s.filter_%v", PackagePrefixResource, resType, op)
	}
	ctx = contextWithOverriddenLogLevel(ctx, res.LogLevel)
	opaOpts, e := PrepareResourcePartialQuery(ctx, res.Query, resType, op, &res)
	if e != nil {
		return nil, ErrInternal.WithMessage(`error when preparing OPA input: %v`, e)
	}

	result, e := res.OPA.Partial(ctx, *opaOpts)
	return handlePartialResult(ctx, result, e)
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

	//if data, e := json.Marshal(opts.Input); e != nil {
	//	eventLogger(ctx, log.LevelError).Printf("Input marshalling error: %v", e)
	//} else {
	//	eventLogger(ctx, log.LevelDebug).Printf("Input: %s", data)
	//}
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

func handlePartialResult(ctx context.Context, result *sdk.PartialResult, rErr error) (_ *sdk.PartialResult, err error) {
	var event partialResultEvent
	defer func() {
		if err == nil {
			eventLogger(ctx, log.LevelDebug).WithKV(kLogPartialResult, event).Printf("Partial [%s]", event.ID)
		} else {
			eventLogger(ctx, log.LevelDebug).WithKV(kLogPartialReason, event).Printf("Deny Partial [%s]", event.ID)
		}
	}()

	if result != nil {
		event.ID = result.ID
	}

	if rErr != nil {
		event.Err = rErr
		switch {
		case sdk.IsUndefinedErr(rErr):
			return nil, ErrAccessDenied
		case errors.Is(rErr, ErrQueriesNotResolved):
			return nil, ErrAccessDenied.WithMessage(rErr.Error())
		default:
			return nil, ErrAccessDenied.WithMessage("failed to perform partial evaluation: %v", rErr)
		}
	}
	event.AST = (*partialQueriesLog)(result.AST)
	return result, nil
}
