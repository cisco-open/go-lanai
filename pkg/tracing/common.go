package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

/**********************
	SpanOptions
 **********************/

func SpanTag(key string, v interface{}) SpanOption {
	return func(span opentracing.Span) {
		span.SetTag(key, v)
	}
}

func SpanBaggageItem(restrictedKey string, s string) SpanOption {
	return func(span opentracing.Span) {
		span.SetBaggageItem(restrictedKey, s)
	}
}

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
