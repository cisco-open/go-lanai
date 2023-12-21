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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	kitopentracing "github.com/go-kit/kit/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/fx"
	"net/http"
)

type httpClientCustomizer struct {
	tracer opentracing.Tracer
}

func HttpClientTracingProvider() fx.Annotated {
	return httpclient.FxClientCustomizers(newHttpClientCustomizer)[0]
}

func newHttpClientCustomizer(tracer opentracing.Tracer) httpclient.ClientCustomizer {
	return &httpClientCustomizer{
		tracer: tracer,
	}
}

func (c *httpClientCustomizer) Customize(opt *httpclient.ClientOption) {
	opt.DefaultBeforeHooks = append(opt.DefaultBeforeHooks,
		httpClientStartSpanHook(c.tracer),
		httpClientTracePropagationHook(c.tracer),
	)
	opt.DefaultAfterHooks = append(opt.DefaultAfterHooks,
		httpClientFinishSpanHook(c.tracer),
	)
}

func httpClientStartSpanHook(tracer opentracing.Tracer) httpclient.BeforeHook {
	fn := func(ctx context.Context, request *http.Request) context.Context {
		name := tracing.OpNameHttpClient + " " + request.Method
		opts := []tracing.SpanOption{
			tracing.SpanKind(ext.SpanKindRPCClientEnum),
			tracing.SpanTag("method", request.Method),
			tracing.SpanTag("url", request.URL.RequestURI()),
		}

		return tracing.WithTracer(tracer).
			WithOpName(name).
			WithOptions(opts...).
			DescendantOrNoSpan(ctx)
	}
	return httpclient.Before(order.Highest, fn)
}

func httpClientTracePropagationHook(tracer opentracing.Tracer) httpclient.BeforeHook {
	fn := func(ctx context.Context, request *http.Request) context.Context {
		reqFunc := kitopentracing.ContextToHTTP(tracer, logger.WithContext(ctx))
		return reqFunc(ctx, request)
	}
	return httpclient.Before(order.Lowest, fn)
}

func httpClientFinishSpanHook(tracer opentracing.Tracer) httpclient.AfterHook {
	fn := func(ctx context.Context, response *http.Response) context.Context {
		op := tracing.WithTracer(tracer).
			WithOptions(tracing.SpanTag("sc", response.StatusCode))
		return op.FinishAndRewind(ctx)
	}
	return httpclient.After(order.Lowest, fn)
}