package tracing

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func MakeBootstrapTracingOption(tracer opentracing.Tracer, opName string) bootstrap.ContextOption {
	return func(ctx context.Context) context.Context {
		return onAppLifecycle(ctx, tracer, opName)
	}
}

func onAppLifecycle(ctx context.Context, tracer opentracing.Tracer, opName string) context.Context {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		span = tracer.StartSpan(opName)
	} else {
		span = tracer.StartSpan(opName, opentracing.ChildOf(span.Context()))
	}
	ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
	return opentracing.ContextWithSpan(ctx, span)
}
