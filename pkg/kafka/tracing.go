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

package kafka

import (
    "context"
    "fmt"
	"github.com/cisco-open/go-lanai/pkg/tracing"
    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "go.uber.org/fx"
)

const tracingOpName = "kafka"

type tracingDI struct {
	fx.In
	Tracer opentracing.Tracer `optional:"true"`
}

func tracingProvider() fx.Annotated {
	return fx.Annotated{
		Group:  FxGroup,
		Target: func(di tracingDI) (ProducerMessageInterceptor, ConsumerDispatchInterceptor, ConsumerHandlerInterceptor) {
			if di.Tracer != nil {
				return newKafkaInterceptors(di.Tracer)
			}
			return nil, nil, nil
		},
	}
}

func newKafkaInterceptors(tracer opentracing.Tracer) (ProducerMessageInterceptor, ConsumerDispatchInterceptor, ConsumerHandlerInterceptor) {
	return &kafkaProducerInterceptor{
			tracer: tracer,
		}, &kafkaConsumerInterceptor{
			tracer: tracer,
		}, &kafkaHandlerInterceptor{
			tracer: tracer,
		}
}

// kafkaProducerInterceptor implements kafka.ProducerMessageInterceptor and kafka.ProducerMessageFinalizer
type kafkaProducerInterceptor struct {
	tracer opentracing.Tracer
}

func (i kafkaProducerInterceptor) Intercept(msgCtx *MessageContext) (*MessageContext, error) {
	cmdStr := "send"
	name := tracingOpName + " " + cmdStr
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCClientEnum),
		tracing.SpanTag("topic", msgCtx.Topic),
		tracing.SpanTag("cmd", cmdStr),
		i.spanPropagation(msgCtx),
	}
	if msgCtx.Key != nil {
		opts = append(opts, tracing.SpanTag("key", fmt.Sprint(msgCtx.Key)))
	}

	ctx := tracing.WithTracer(i.tracer).
		WithOpName(name).
		WithOptions(opts...).
		FollowsOrNoSpan(msgCtx.Context)

	msgCtx.Context = ctx
	return msgCtx, nil
}

func (i kafkaProducerInterceptor) Finalize(msgCtx *MessageContext, p int32, offset int64, err error) (*MessageContext, error) {
	op := tracing.WithTracer(i.tracer)
	if err != nil {
		op = op.WithOptions(tracing.SpanTag("err", err))
	} else {
		op = op.
			WithOptions(tracing.SpanTag("partition", p)).
			WithOptions(tracing.SpanTag("offset", offset))
	}
	msgCtx.Context = op.FinishAndRewind(msgCtx.Context)
	return msgCtx, err
}

func (i kafkaProducerInterceptor) spanPropagation(msgCtx *MessageContext) tracing.SpanOption {
	return func(span opentracing.Span) {
		// we ignore error, since we can't do anything about it
		_ = i.tracer.Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(msgCtx.Message.Headers))
	}
}

// kafkaProducerInterceptor implements kafka.ConsumerDispatchInterceptor and kafka.ConsumerDispatchFinalizer
type kafkaConsumerInterceptor struct {
	tracer opentracing.Tracer
}

func (i kafkaConsumerInterceptor) Intercept(msgCtx *MessageContext) (*MessageContext, error) {

	// first extract span from message
	ctx := tracing.WithTracer(i.tracer).
		WithStartOptions(i.spanPropagation(msgCtx)).
		ForceNewSpan(msgCtx.Context)

	// second, start a follower span
	cmdStr := "recv"
	switch msgCtx.Source.(type) {
	case Subscriber:
		cmdStr = "subscribe"
	case GroupConsumer:
		cmdStr = "consume"
	}
	name := tracingOpName + " " + cmdStr
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCServerEnum),
		tracing.SpanTag("topic", msgCtx.Topic),
		tracing.SpanTag("cmd", cmdStr),
	}
	if msgCtx.Key != nil {
		opts = append(opts, tracing.SpanTag("key", fmt.Sprint(msgCtx.Key)))
	}

	ctx = tracing.WithTracer(i.tracer).
		WithOpName(name).
		WithOptions(opts...).
		NewSpanOrFollows(ctx)

	msgCtx.Context = ctx
	return msgCtx, nil
}

func (i kafkaConsumerInterceptor) Finalize(msgCtx *MessageContext, err error) (*MessageContext, error) {
	op := tracing.WithTracer(i.tracer)
	if err != nil {
		op = op.WithOptions(tracing.SpanTag("err", err))
	}
	msgCtx.Context = op.FinishAndRewind(msgCtx.Context)
	return msgCtx, err
}

func (i kafkaConsumerInterceptor) spanPropagation(msgCtx *MessageContext) opentracing.StartSpanOption {
	spanCtx, e := i.tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(msgCtx.Message.Headers))
	if e != nil {
		return noopStartSpanOption{}
	}
	return ext.RPCServerOption(spanCtx)
}

// kafkaProducerInterceptor implements kafka.ConsumerHandlerInterceptor
type kafkaHandlerInterceptor struct {
	tracer opentracing.Tracer
}

func (i kafkaHandlerInterceptor) BeforeHandling(ctx context.Context, _ *Message) (context.Context, error) {
	cmdStr := "handle"
	name := tracingOpName + " " + cmdStr
	opts := []tracing.SpanOption{
		tracing.SpanKind(ext.SpanKindRPCServerEnum),
		tracing.SpanTag("cmd", cmdStr),
	}

	ctx = tracing.WithTracer(i.tracer).
		WithOpName(name).
		WithOptions(opts...).
		DescendantOrNoSpan(ctx)

	return ctx, nil
}

func (i kafkaHandlerInterceptor) AfterHandling(ctx context.Context, _ *Message, err error) (context.Context, error) {
	op := tracing.WithTracer(i.tracer)
	if err != nil {
		op = op.WithOptions(tracing.SpanTag("err", err))
	}
	ctx = op.FinishAndRewind(ctx)
	return ctx, err
}

type noopStartSpanOption struct{}

func (o noopStartSpanOption) Apply(*opentracing.StartSpanOptions){}
