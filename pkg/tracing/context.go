package tracing

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
	"time"
)

var logger = log.New("Tracing")

const (
	OpNameBootstrap = "bootstrap"
	OpNameStart = "startup"
	OpNameStop = "shutdown"
	OpNameHttp = "http"
	OpNameRedis = "redis"
	//OpName = ""
)

const (
	spanFinisherKey = "SF"
)

type SpanOption func(opentracing.Span)

type SpanRewinder func() context.Context

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

func SpanRewinderFromContext(ctx context.Context) SpanRewinder {
	if finisher, ok := ctx.Value(spanFinisherKey).(SpanRewinder); ok {
		return finisher
	}

	// try to get from Request's context if given context contains gin.Context
	if gc := web.GinContext(ctx); gc != nil {
		if finisher, ok := gc.Request.Context().Value(spanFinisherKey).(SpanRewinder); ok {
			return finisher
		}
	}
	return nil
}

func ContextWithSpanRewinder(ctx context.Context, finisher SpanRewinder) context.Context {
	return context.WithValue(ctx, spanFinisherKey, finisher)
}

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
func (op SpanOperator) UpdateCurrentSpan(ctx context.Context)  {
	span := SpanFromContext(ctx)
	if span == nil {
		return
	}
	op.applyUpdateOptions(span)
	return
}

// Finish finish current span if exist.
// Note: The finished span is still counted as "current span".
//		 If caller want to rewind to previous span, use FinishAndRewind instead
func (op SpanOperator) Finish(ctx context.Context) {
	if span := SpanFromContext(ctx); span != nil {
		op.applyUpdateOptions(span)
		op.finishOptions.FinishTime = time.Now().UTC()
		span.FinishWithOptions(op.finishOptions)
	}
}

// FinishAndRewind finish current span if exist and restore context with parent span if possible (no garantees)
// callers shall not continue to use the old context after this call
// Note: all values in given context added during the current span will be lost. It's like rewind operation
func (op SpanOperator) FinishAndRewind(ctx context.Context) context.Context {
	op.Finish(ctx)
	rewinder := SpanRewinderFromContext(ctx)
	if rewinder == nil {
		return ctx
	}
	return rewinder()
}

// NewSpanOrDescendant create new span if not currently have one,
// spawn a child span using opentracing.ChildOf(span.Context()) if span exists
func (op SpanOperator) NewSpanOrDescendant(ctx context.Context) context.Context {
	return op.newSpan(ctx, func(span opentracing.Span) opentracing.SpanReference {
		return opentracing.ChildOf(span.Context())
	}, true)
}

// NewSpanOrDescendant create new span if not currently have one,
// spawn a child span using opentracing.FollowsFrom(span.Context()) if span exists
func (op SpanOperator) NewSpanOrFollows(ctx context.Context) context.Context {
	return op.newSpan(ctx, func(span opentracing.Span) opentracing.SpanReference {
		return opentracing.FollowsFrom(span.Context())
	}, true)
}

// DescendantOrNoSpan spawn a child span using opentracing.ChildOf(span.Context()) if there is a span exists
// otherwise do nothing
func (op SpanOperator) DescendantOrNoSpan(ctx context.Context) context.Context {
	return op.newSpan(ctx, func(span opentracing.Span) opentracing.SpanReference {
		return opentracing.ChildOf(span.Context())
	}, false)
}

// NewSpanOrDescendant spawn a child span using opentracing.FollowsFrom(span.Context()) if there is a span exists
// otherwise do nothing
func (op SpanOperator) FollowsOrNoSpan(ctx context.Context) context.Context {
	return op.newSpan(ctx, func(span opentracing.Span) opentracing.SpanReference {
		return opentracing.FollowsFrom(span.Context())
	}, false)
}

func (op SpanOperator) createSpanRewinder(ctx context.Context) SpanRewinder {
	return func() context.Context {
		return ctx
	}
}

type spanReferencer func(opentracing.Span) opentracing.SpanReference

func (op *SpanOperator) newSpan(ctx context.Context, referencer spanReferencer, must bool) context.Context {
	span := SpanFromContext(ctx)
	rewinder := op.createSpanRewinder(ctx)
	switch {
	case span != nil:
		options := append([]opentracing.StartSpanOption{referencer(span)}, op.startOptions...)
		span = op.tracer.StartSpan(op.name, options...)
	case must:
		span = op.tracer.StartSpan(op.name, op.startOptions...)
	default:
		return ctx
	}

	op.applyUpdateOptions(span)
	return opentracing.ContextWithSpan(ContextWithSpanRewinder(ctx, rewinder), span)
}

func (op SpanOperator) applyUpdateOptions(span opentracing.Span) {
	for _, ext := range op.updateOptions {
		ext(span)
	}
}


