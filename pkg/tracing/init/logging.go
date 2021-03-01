package tracing

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"fmt"
	"github.com/uber/jaeger-client-go"
)

var tracingLogValuers = log.ContextValuers{
	"traceId": traceIdContextValuer,
	"spanId": spanIdContextValuer,
	"parentId": parentIdContextValuer,
}

func traceIdContextValuer(ctx context.Context) (ret interface{}) {
	span := tracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	switch span.Context().(type) {
	case jaeger.SpanContext:
		ret = jaegerTraceIdString(span.Context().(jaeger.SpanContext).TraceID())
	default:
		return
	}
	return
}

func spanIdContextValuer(ctx context.Context) (ret interface{}) {
	span := tracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	switch span.Context().(type) {
	case jaeger.SpanContext:
		ret = jaegerSpanIdString(span.Context().(jaeger.SpanContext).SpanID())
	}
	return
}

func parentIdContextValuer(ctx context.Context) (ret interface{}) {
	span := tracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	switch span.Context().(type) {
	case jaeger.SpanContext:
		ret = jaegerSpanIdString(span.Context().(jaeger.SpanContext).ParentID())
	default:
		return
	}
	return
}

func jaegerTraceIdString(id jaeger.TraceID) string {
	if !id.IsValid() {
		return ""
	}
	if id.High == 0 {
		return fmt.Sprintf("%.16x", id.Low)
	}
	return fmt.Sprintf("%.16x%016x", id.High, id.Low)
}

func jaegerSpanIdString(id jaeger.SpanID) string {
	if id != 0 {
		return fmt.Sprintf("%.16x", uint64(id))
	}
	return ""
}