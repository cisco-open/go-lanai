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

package opensearch

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/fx"
)

// Tracer will provide some opensearch.HookContainer to provide tracing
type Tracer struct {
	tracer opentracing.Tracer
}

func (t *Tracer) Before(ctx context.Context, before BeforeContext) context.Context {
	if t.tracer == nil {
		return ctx
	}
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCClientEnum),
		tracing.SpanTag("command", before.CommandType()),
	}
	ctx = tracing.WithTracer(t.tracer).
		WithOpName("opensearch " + before.CommandType().String()).
		WithOptions(opts...).
		DescendantOrNoSpan(ctx)
	return ctx
}

func (t *Tracer) After(ctx context.Context, afterContext AfterContext) context.Context {
	if t.tracer == nil {
		return ctx
	}
	op := tracing.WithTracer(t.tracer)

	if (afterContext.Resp) != nil && (afterContext.Resp).IsError() {
		op = op.WithOptions(
			tracing.SpanTag("status code", (afterContext.Resp).StatusCode),
		)
	} else if *afterContext.Err != nil {
		op = op.WithOptions(
			tracing.SpanTag("error", afterContext.Err),
		)
	} else {
		if afterContext.CommandType() == CmdSearch {
			resp, err := UnmarshalResponse[SearchResponse[any]](afterContext.Resp)
			if err != nil {
				logger.Errorf("unable to unmarshal error: %v", err)
			} else {
				op = op.WithOptions(
					tracing.SpanTag("hits", resp.Hits.Total.Value),
					tracing.SpanTag("maxscore", resp.Hits.MaxScore),
				)
			}
		}
	}

	ctx = op.FinishAndRewind(ctx)
	return ctx
}

func TracerHook(tracer opentracing.Tracer) *Tracer {
	o := Tracer{
		tracer: tracer,
	}
	return &o
}

type tracingDI struct {
	fx.In
	Tracer opentracing.Tracer `optional:"true"`
}

func tracingProvider() fx.Annotated {
	return fx.Annotated{
		Group: FxGroup,
		Target: func(di tracingDI) (BeforeHook, AfterHook) {
			tracerHook := TracerHook(di.Tracer)
			return tracerHook, tracerHook
		},
	}
}
