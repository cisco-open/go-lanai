package instrument

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/kafka"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/fx"
)

// kafkaProducerInterceptor implements kafka.ProducerInterceptor and kafka.ProducerMessageFinalizer
type kafkaProducerInterceptor struct {
	tracer opentracing.Tracer
}

func KafkaTracingTracingProvider() fx.Annotated {
	return fx.Annotated{
		Group:  kafka.FxGroup,
		Target: newKafkaProducerInterceptor,
	}
}

func newKafkaProducerInterceptor(tracer opentracing.Tracer) kafka.ProducerInterceptor {
	return &kafkaProducerInterceptor{
		tracer: tracer,
	}
}

func (i kafkaProducerInterceptor) Intercept(msgCtx *kafka.MessageContext) (*kafka.MessageContext, error) {
	cmdStr := "send"
	name := tracing.OpNameKafka + " " + cmdStr
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

	logger.WithContext(ctx).Debugf("Traced kafka message [->%s]: %v", msgCtx.Topic, msgCtx.Payload)
	msgCtx.Context = ctx
	return msgCtx, nil
}

func (i kafkaProducerInterceptor) Finalize(msgCtx *kafka.MessageContext, p int32, offset int64, err error) (*kafka.MessageContext, error) {
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

func (i kafkaProducerInterceptor) spanPropagation(msgCtx *kafka.MessageContext) tracing.SpanOption {
	return func(span opentracing.Span) {
		// we ignore error, since we can't do anything about it
		_ = i.tracer.Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(msgCtx.Headers))
	}
}