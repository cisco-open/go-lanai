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

package tracing

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
	"time"
)

type spanKey struct{}

var spanFinisherKey = spanKey{}

// DefaultLogValuers is used by log package to extract tracing information in log templates.
// This variable is properly set by "tracing/init".
var DefaultLogValuers = LogValuers{
	TraceIDValuer:  func(context.Context) interface{} { return nil },
	SpanIDValuer:   func(context.Context) interface{} { return nil },
	ParentIDValuer: func(context.Context) interface{} { return nil },
}

type SpanOption func(opentracing.Span)

type SpanRewinder func() context.Context

/**********************
	Context
 **********************/

type LogValuers struct {
	TraceIDValuer  log.ContextValuer
	SpanIDValuer   log.ContextValuer
	ParentIDValuer log.ContextValuer
}

func (v LogValuers) ContextValuers() log.ContextValuers {
	return log.ContextValuers{
		"traceId":  v.TraceIDValuer,
		"spanId":   v.SpanIDValuer,
		"parentId": v.ParentIDValuer,
	}
}

//nolint:contextcheck
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

func TraceIdFromContext(ctx context.Context) (ret interface{}) {
	return DefaultLogValuers.TraceIDValuer(ctx)
}

func SpanIdFromContext(ctx context.Context) (ret interface{}) {
	return DefaultLogValuers.SpanIDValuer(ctx)
}

func ParentIdFromContext(ctx context.Context) (ret interface{}) {
	return DefaultLogValuers.ParentIDValuer(ctx)
}

/**********************
	Span Operators
 **********************/

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

func (op *SpanOperator) UpdateCurrentSpan(ctx context.Context) {
	span := SpanFromContext(ctx)
	if span == nil {
		return
	}
	op.applyUpdateOptions(span)
	return
}

// Finish finish current span if exist.
// Note: The finished span is still counted as "current span".
//
//	If caller want to rewind to previous span, use FinishAndRewind instead
func (op *SpanOperator) Finish(ctx context.Context) {
	if span := SpanFromContext(ctx); span != nil {
		op.applyUpdateOptions(span)
		op.finishOptions.FinishTime = time.Now().UTC()
		span.FinishWithOptions(op.finishOptions)
	}
}

// FinishAndRewind finish current span if exist and restore context with parent span if possible (no garantees)
// callers shall not continue to use the old context after this call
// Note: all values in given context added during the current span will be lost. It's like rewind operation
func (op *SpanOperator) FinishAndRewind(ctx context.Context) context.Context {
	op.Finish(ctx)
	rewinder := SpanRewinderFromContext(ctx)
	if rewinder == nil {
		return ctx
	}
	return rewinder()
}

// NewSpanOrDescendant create new span if not currently have one,
// spawn a child span using opentracing.ChildOf(span.Context()) if span exists
func (op *SpanOperator) NewSpanOrDescendant(ctx context.Context) context.Context {
	return op.newSpan(ctx, func(span opentracing.Span) opentracing.SpanReference {
		return opentracing.ChildOf(span.Context())
	}, true)
}

// NewSpanOrFollows create new span if not currently have one,
// spawn a child span using opentracing.FollowsFrom(span.Context()) if span exists
func (op *SpanOperator) NewSpanOrFollows(ctx context.Context) context.Context {
	return op.newSpan(ctx, func(span opentracing.Span) opentracing.SpanReference {
		return opentracing.FollowsFrom(span.Context())
	}, true)
}

// DescendantOrNoSpan spawn a child span using opentracing.ChildOf(span.Context()) if there is a span exists
// otherwise do nothing
func (op *SpanOperator) DescendantOrNoSpan(ctx context.Context) context.Context {
	return op.newSpan(ctx, func(span opentracing.Span) opentracing.SpanReference {
		return opentracing.ChildOf(span.Context())
	}, false)
}

// FollowsOrNoSpan spawn a child span using opentracing.FollowsFrom(span.Context()) if there is a span exists
// otherwise do nothing
func (op *SpanOperator) FollowsOrNoSpan(ctx context.Context) context.Context {
	return op.newSpan(ctx, func(span opentracing.Span) opentracing.SpanReference {
		return opentracing.FollowsFrom(span.Context())
	}, false)
}

// ForceNewSpan force to create a new span and discard any existing span
// Warning: Internal usage, use with caution
func (op *SpanOperator) ForceNewSpan(ctx context.Context) context.Context {
	return op.newSpan(ctx, nil, true)
}

func (op *SpanOperator) createSpanRewinder(ctx context.Context) SpanRewinder {
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
		options := op.startOptions
		if referencer != nil {
			options = append([]opentracing.StartSpanOption{referencer(span)}, options...)
		}
		span = op.tracer.StartSpan(op.name, options...)
	case must:
		span = op.tracer.StartSpan(op.name, op.startOptions...)
	default:
		return ctx
	}

	op.applyUpdateOptions(span)
	return opentracing.ContextWithSpan(ContextWithSpanRewinder(ctx, rewinder), span)
}

func (op *SpanOperator) applyUpdateOptions(span opentracing.Span) {
	for _, ext := range op.updateOptions {
		ext(span)
	}
}
