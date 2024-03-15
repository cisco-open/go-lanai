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

package webtracing

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/http"
)

var (
	healthMatcher = matcher.RequestWithPattern("**/health")
	corsPreflightMatcher = matcher.RequestWithMethods(http.MethodOptions)
	excludeRequest = corsPreflightMatcher.Or(healthMatcher)
)

type tracingWebCustomizer struct {
	tracer opentracing.Tracer
}

func newTracingWebCustomizer(tracer opentracing.Tracer) *tracingWebCustomizer {
	return &tracingWebCustomizer{
		tracer: tracer,
	}
}

// Order we want tracingWebCustomizer before anything else
func (c tracingWebCustomizer) Order() int {
	return order.Highest
}

func (c tracingWebCustomizer) Customize(_ context.Context, r *web.Registrar) error {
	//nolint:contextcheck // false positive
	if e := r.AddGlobalMiddlewares(GinTracing(c.tracer, tracing.OpNameHttp, excludeRequest)); e != nil {
		return e
	}
	return nil
}

func GinTracing(tracer opentracing.Tracer, opName string, excludes web.RequestMatcher) gin.HandlerFunc {
	return func(gc *gin.Context) {
		if m, e := excludes.Matches(gc.Request); e == nil && m {
			return
		}

		// start or join span
		orig := gc.Request.Context()
		ctx := contextWithRequest(orig, tracer, gc.Request, opName)
		gc.Request = gc.Request.WithContext(ctx)

		gc.Next()

		// finish the span
		tracing.WithTracer(tracer).
			WithOptions(tracing.SpanHttpStatusCode(gc.Writer.Status())).
			Finish(ctx)
		gc.Request = gc.Request.WithContext(orig)
	}
}

/*********************
	common funcs
 *********************/

func opNameWithRequest(opName string, r *http.Request) string {
	return opName + " " + r.URL.Path
}

func contextWithRequest(ctx context.Context, tracer opentracing.Tracer, req *http.Request, opName string) context.Context {
	opName = opNameWithRequest(opName, req)
	spanOp := tracing.WithTracer(tracer).
		WithOpName(opName).
		WithOptions(
			tracing.SpanKind(ext.SpanKindRPCServerEnum),
			tracing.SpanHttpMethod(req.Method),
			tracing.SpanHttpUrl(req.URL.String()),
		)

	if spanCtx, e := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)); e == nil {
		spanOp = spanOp.WithStartOptions(ext.RPCServerOption(spanCtx))
	}
	return spanOp.NewSpanOrDescendant(ctx)
}
