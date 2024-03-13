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

package vaulttracing

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const OpName = "vault"

type Hook struct {
	tracer opentracing.Tracer
}

func NewHook(tracer opentracing.Tracer) *Hook {
	return &Hook{
		tracer: tracer,
	}
}

func (h *Hook) BeforeOperation(ctx context.Context, cmd string) context.Context {
	name := OpName + " " + cmd
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCClientEnum),
		tracing.SpanTag("cmd", cmd),
	}
	return tracing.WithTracer(h.tracer).
		WithOpName(name).
		WithOptions(opts...).
		DescendantOrNoSpan(ctx)
}

func (h *Hook) AfterOperation(ctx context.Context, err error) {
	op := tracing.WithTracer(h.tracer)
	if err != nil {
		op.WithOptions(tracing.SpanTag("err", err))
	}
	op.Finish(ctx)
}
