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
	"time"
)

type QueryOptions func(q *Query)

type Query struct {
	OPA              *sdk.OPA
	Policy           string
	InputCustomizers []InputCustomizer
	RawInput         interface{}
	// LogLevel override decision log level when presented
	LogLevel *log.LoggingLevel
}

func QueryWithPolicy(policy string) QueryOptions {
	return func(q *Query) {
		q.Policy = policy
	}
}

func SilentQuery() QueryOptions {
	var silent = log.LevelOff
	return func(opt *Query) {
		opt.LogLevel = &silent
	}
}

func QueryWithInputCustomizer(customizer InputCustomizerFunc) QueryOptions {
	return func(q *Query) {
		q.InputCustomizers = append(q.InputCustomizers, customizer)
	}
}

// Allow is generic API for querying policy. This function only populate minimum input data like authentication.
// For more specialized function, see AllowResource, AllowRequest, etc.
func Allow(ctx context.Context, opts ...QueryOptions) error {
	query := Query{
		OPA:              EmbeddedOPA(),
		InputCustomizers: embeddedOPA.inputCustomizers,
	}
	for _, fn := range opts {
		fn(&query)
	}
	if len(query.Policy) == 0 {
		return ErrInternal.WithMessage("policy is required for generic Allow function")
	}
	ctx = contextWithOverriddenLogLevel(ctx, query.LogLevel)
	opaOpts, e := PrepareGenericDecisionQuery(ctx, &query)
	if e != nil {
		return ErrInternal.WithMessage(`error when preparing OPA input: %v`, e)
	}
	result, e := query.OPA.Decision(ctx, *opaOpts)
	return handleDecisionResult(ctx, result, e, "")
}

func PrepareGenericDecisionQuery(ctx context.Context, query *Query) (*sdk.DecisionOptions, error) {
	input, e := constructGenericDecisionInput(ctx, query)
	if e != nil {
		return nil, e
	}
	opts := sdk.DecisionOptions{
		Now:                 time.Now(),
		Path:                query.Policy,
		Input:               input,
		StrictBuiltinErrors: false,
	}

	//if data, e := json.Marshal(opts.Input); e != nil {
	//	eventLogger(ctx, log.LevelError).Printf("Input marshalling error: %v", e)
	//} else {
	//	eventLogger(ctx, log.LevelDebug).Printf("Input: %s", data)
	//}
	return &opts, nil
}

func constructGenericDecisionInput(ctx context.Context, query *Query) (interface{}, error) {
	if query.RawInput != nil {
		return query.RawInput, nil
	}
	input := NewInput()
	input.Authentication = NewAuthenticationClause()
	for _, customizer := range query.InputCustomizers {
		if e := customizer.Customize(ctx, input); e != nil {
			return nil, e
		}
	}
	return input, nil
}
