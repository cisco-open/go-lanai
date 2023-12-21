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

package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opensearch"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/fx"
)

// OpensearchTracer will provide some opensearch.HookContainer to provide tracing
type OpenSearchTracer struct {
	tracer opentracing.Tracer
}

func (o *OpenSearchTracer) Before(ctx context.Context, before opensearch.BeforeContext) context.Context {
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCClientEnum),
		tracing.SpanTag("command", before.CommandType()),
	}
	ctx = tracing.WithTracer(o.tracer).
		WithOpName("opensearch " + before.CommandType().String()).
		WithOptions(opts...).
		DescendantOrNoSpan(ctx)
	return ctx
}

func (o *OpenSearchTracer) After(ctx context.Context, afterContext opensearch.AfterContext) context.Context {
	op := tracing.WithTracer(o.tracer)

	if (afterContext.Resp) != nil && (afterContext.Resp).IsError() {
		op = op.WithOptions(
			tracing.SpanTag("status code", (afterContext.Resp).StatusCode),
		)
	} else if *afterContext.Err != nil {
		op = op.WithOptions(
			tracing.SpanTag("error", afterContext.Err),
		)
	} else {
		if afterContext.CommandType() == opensearch.CmdSearch {
			resp, err := opensearch.UnmarshalResponse[opensearch.SearchResponse[any]](afterContext.Resp)
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

func OpenSearchTracerHook(tracer opentracing.Tracer) *OpenSearchTracer {
	o := OpenSearchTracer{
		tracer: tracer,
	}
	return &o
}

func OpenSearchTracingProvider() fx.Annotated {
	return fx.Annotated{
		Group: opensearch.FxGroup,
		Target: func(tracer opentracing.Tracer) (opensearch.BeforeHook, opensearch.AfterHook) {
			tracerHook := OpenSearchTracerHook(tracer)
			return tracerHook, tracerHook
		},
	}
}
