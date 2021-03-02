package tracing

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
	"time"
)

const (
	OpNameBootstrap = "bootstrap"
	OpNameStart = "startup"
	OpNameStop = "shutdown"
	OpNameHttp = "http"
	OpNameRedis = "redis"
	//OpName = ""
)

func SpanFromContext(ctx context.Context) (span opentracing.Span) {
	span = opentracing.SpanFromContext(ctx)
	if span != nil {
		return
	}

	// try to get from Request's context if given context contains gin.Context
	if gc := web.GinContext(ctx); gc != nil {
		span = opentracing.SpanFromContext(gc.Request.Context())
	}
	return
}

type SpanOption func(opentracing.Span)

type SpanOperator struct {
	tracer        opentracing.Tracer
	name          string
	startOptions  []opentracing.StartSpanOption
	updateOptions []SpanOption
	finishOptions opentracing.FinishOptions
}

func WithTracer(tracer opentracing.Tracer) *SpanOperator {
	return &SpanOperator{
		tracer:        tracer,
		startOptions:  []opentracing.StartSpanOption{},
		updateOptions: []SpanOption{},
	}
}

// Setters
func (op *SpanOperator) WithOpName(name string) *SpanOperator {
	op.name = name
	return op
}

func (op *SpanOperator) WithStartOptions(options ...opentracing.StartSpanOption) *SpanOperator {
	op.startOptions = append(op.startOptions, options...)
	return op
}

func (op *SpanOperator) WithOptions(exts ...SpanOption) *SpanOperator {
	op.updateOptions = append(op.updateOptions, exts...)
	return op
}

// Operations
func (op SpanOperator) applyUpdateOptions(span opentracing.Span) {
	for _, ext := range op.updateOptions {
		ext(span)
	}
}

func (op SpanOperator) UpdateCurrentSpan(ctx context.Context) {
	span := SpanFromContext(ctx)
	if span == nil {
		return
	}
	op.applyUpdateOptions(span)
	return
}

func (op SpanOperator) FinishCurrentSpan(ctx context.Context) {
	span := SpanFromContext(ctx)
	if span == nil {
		return
	}
	op.applyUpdateOptions(span)
	op.finishOptions.FinishTime = time.Now().UTC()
	span.FinishWithOptions(op.finishOptions)
	return
}

func (op SpanOperator) NewSpanOrDescendant(ctx context.Context) context.Context {
	span := SpanFromContext(ctx)
	if span == nil {
		span = op.tracer.StartSpan(op.name, op.startOptions...)
	} else {
		opentracing.ChildOf(span.Context())
		options := append([]opentracing.StartSpanOption{
			opentracing.ChildOf(span.Context()),
		}, op.startOptions...)
		span = op.tracer.StartSpan(op.name, options...)
	}
	op.applyUpdateOptions(span)
	return opentracing.ContextWithSpan(ctx, span)
}

func (op SpanOperator) NewSpanOrFollows(ctx context.Context) context.Context {
	span := SpanFromContext(ctx)
	if span == nil {
		span = op.tracer.StartSpan(op.name, op.startOptions...)
	} else {
		opentracing.ChildOf(span.Context())
		options := append([]opentracing.StartSpanOption{
			opentracing.FollowsFrom(span.Context()),
		}, op.startOptions...)
		span = op.tracer.StartSpan(op.name, options...)
	}
	op.applyUpdateOptions(span)
	return opentracing.ContextWithSpan(ctx, span)
}

