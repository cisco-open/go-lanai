package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type TracingTaskHook struct {
	tracer opentracing.Tracer
}

func NewTracingTaskHook(tracer opentracing.Tracer) *TracingTaskHook {
	return &TracingTaskHook{
		tracer: tracer,
	}
}

func (h *TracingTaskHook) BeforeTrigger(ctx context.Context, id string) context.Context {
	name := tracing.OpNameScheduler
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCClientEnum),
		tracing.SpanTag("task", id),
	}

	return tracing.WithTracer(h.tracer).
		WithOpName(name).
		WithOptions(opts...).
		NewSpanOrFollows(ctx)
}

func (h *TracingTaskHook) AfterTrigger(ctx context.Context, _ string, err error) {
	op := tracing.WithTracer(h.tracer)
	if err != nil {
		op.WithOptions(tracing.SpanTag("err", err))
	}
	op.Finish(ctx)
	return
}
