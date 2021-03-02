package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func MakeBootstrapTracingOption(tracer opentracing.Tracer, opName string) bootstrap.ContextOption {
	return func(ctx context.Context) context.Context {
		return tracing.WithTracer(tracer).
			WithOpName(opName).
			WithOptions(tracing.SpanKind(ext.SpanKindRPCServerEnum)).
			NewSpanOrDescendant(ctx)
	}
}

func MakeStartTracingOption(tracer opentracing.Tracer, opName string) bootstrap.ContextOption {
	return func(ctx context.Context) context.Context {
		return tracing.WithTracer(tracer).
			WithOpName(opName).
			WithOptions(tracing.SpanKind(ext.SpanKindRPCServerEnum)).
			NewSpanOrDescendant(ctx)
	}
}

func MakeStopTracingOption(tracer opentracing.Tracer, opName string) bootstrap.ContextOption {
	return func(ctx context.Context) context.Context {
		// finish current if not root span and start a new child
		return tracing.WithTracer(tracer).
			WithOpName(opName).
			WithOptions(tracing.SpanKind(ext.SpanKindRPCServerEnum)).
			NewSpanOrDescendant(ctx)
	}
}

