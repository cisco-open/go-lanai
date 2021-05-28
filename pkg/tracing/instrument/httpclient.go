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