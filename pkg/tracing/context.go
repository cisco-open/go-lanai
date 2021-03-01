package tracing

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
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
