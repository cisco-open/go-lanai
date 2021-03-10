package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type vaultTracingHook struct {
	tracer opentracing.Tracer
}

func NewVaultTracingHook(tracer opentracing.Tracer) *vaultTracingHook {
	return &vaultTracingHook{
		tracer: tracer,
	}
}

func (v *vaultTracingHook) BeforeOperation(ctx context.Context, cmd string) context.Context {
	name := tracing.OpNameVault + " " + cmd
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCClientEnum),
		tracing.SpanTag("cmd", cmd),
	}
	return tracing.WithTracer(v.tracer).
		WithOpName(name).
		WithOptions(opts...).
		DescendantOrNoSpan(ctx)
}

func (v *vaultTracingHook) AfterOperation(ctx context.Context, err error)  {
	op := tracing.WithTracer(v.tracer)
	if err != nil {
		op.WithOptions(tracing.SpanTag("err", err))
	}
	op.Finish(ctx)
}
