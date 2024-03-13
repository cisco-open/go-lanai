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

package httpclient

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/fx"
	"net"
	"net/http"
	"strconv"
)

const TracingOpName = "remote-http"

type tracingCustomizer struct {
	tracer opentracing.Tracer
}

func tracingProvider() fx.Annotated {
	return FxClientCustomizers(newTracingCustomizer)[0]
}

type tracingDI struct {
	fx.In
	Tracer opentracing.Tracer `optional:"true"`
}

func newTracingCustomizer(di tracingDI) ClientCustomizer {
	return &tracingCustomizer{
		tracer: di.Tracer,
	}
}

func (c *tracingCustomizer) Customize(opt *ClientOption) {
	if c.tracer == nil {
		return
	}
	opt.DefaultBeforeHooks = append(opt.DefaultBeforeHooks,
		startSpanHook(c.tracer),
	)
	opt.DefaultAfterHooks = append(opt.DefaultAfterHooks,
		finishSpanHook(c.tracer),
	)
}

func startSpanHook(tracer opentracing.Tracer) BeforeHook {
	fn := func(ctx context.Context, req *http.Request) context.Context {
		name := TracingOpName + " " + req.Method

		opts := []tracing.SpanOption{
			tracing.SpanKind(ext.SpanKindRPCClientEnum),
			tracing.SpanTag("method", req.Method),
			tracing.SpanTag("url", req.URL.RequestURI()),

		}

		// standard tags
		hostname := req.URL.Host
		var port int
		if host, portString, e := net.SplitHostPort(req.URL.Host); e == nil {
			hostname = host
			port, _ = strconv.Atoi(portString)
		}
		opts = append(opts,
			tracing.SpanHttpMethod(req.Method),
			tracing.SpanHttpUrl(req.URL.String()),
			func(span opentracing.Span) {
				ext.PeerHostname.Set(span, hostname)
				if port != 0 {
					ext.PeerPort.Set(span, uint16(port))
				}
			},
		)

		// propagation
		opts = append(opts, spanPropagation(req, tracer))

		return tracing.WithTracer(tracer).
			WithOpName(name).
			WithOptions(opts...).
			DescendantOrNoSpan(ctx)
	}
	return BeforeHookWithOrder(order.Highest, BeforeHookFunc(fn))
}

func finishSpanHook(tracer opentracing.Tracer) AfterHook {
	fn := func(ctx context.Context, response *http.Response) context.Context {
		op := tracing.WithTracer(tracer).
			WithOptions(
				tracing.SpanTag("sc", response.StatusCode),
				tracing.SpanHttpStatusCode(response.StatusCode),
			)
		return op.FinishAndRewind(ctx)
	}
	return AfterHookWithOrder(order.Lowest, AfterHookFunc(fn))
}

func spanPropagation(req *http.Request, tracer opentracing.Tracer) tracing.SpanOption {
	return func(span opentracing.Span) {
		_ = tracer.Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header),
		)
	}
}
