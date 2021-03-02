package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func MakeLifecycleTracingOption(tracer opentracing.Tracer, opName string) bootstrap.ContextOption {
	return func(ctx context.Context) context.Context {
		return onAppLifecycle(ctx, tracer, opName)
	}
}

func onAppLifecycle(ctx context.Context, tracer opentracing.Tracer, opName string) context.Context {
	return tracing.WithTracer(tracer).
		WithOpName(opName).
		WithOptions(tracing.SpanKind(ext.SpanKindRPCServerEnum)).
		NewSpanOrDescendant(ctx)
}
