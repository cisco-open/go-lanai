package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func SpanKind(v ext.SpanKindEnum) SpanOption {
	return func(span opentracing.Span) {
		ext.SpanKind.Set(span, v)
	}
}

func SpanComponent(v string) SpanOption {
	return func(span opentracing.Span) {
		ext.Component.Set(span, v)
	}
}

func SpanHttpUrl(v string) SpanOption {
	return func(span opentracing.Span) {
		ext.HTTPUrl.Set(span, v)
	}
}

func SpanHttpMethod(v string) SpanOption {
	return func(span opentracing.Span) {
		ext.HTTPMethod.Set(span, v)
	}
}

func SpanHttpStatusCode(v int) SpanOption {
	return func(span opentracing.Span) {
		ext.HTTPStatusCode.Set(span, uint16(v))
	}
}
