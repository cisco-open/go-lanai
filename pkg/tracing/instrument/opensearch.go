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

type TracingProvider struct {
	fx.Out

	Before opensearch.BeforeHook `group:"opensearch_before_hooks"`
	After  opensearch.AfterHook  `group:"opensearch_after_hooks"`
}

func OpenSearchTracingProvider(tracer opentracing.Tracer) TracingProvider {
	tracerHook := OpenSearchTracerHook(tracer)
	return TracingProvider{
		Before: tracerHook,
		After:  tracerHook,
	}
}
