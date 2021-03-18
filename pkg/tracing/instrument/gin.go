package instrument

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	kitopentracing "github.com/go-kit/kit/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
)

func GinTracing(tracer opentracing.Tracer, opName string, excludes web.RequestMatcher) gin.HandlerFunc {
	return func(gc *gin.Context) {
		if m, e := excludes.Matches(gc.Request); e == nil && m {
			return
		}

		orig := gc.Request.Context()

		// start or join span
		reqFunc := kitopentracing.HTTPToContext(tracer, opNameWithRequest(opName, gc.Request), logger)
		ctx := reqFunc(orig, gc.Request)
		gc.Request = gc.Request.WithContext(ctx)

		gc.Next()

		// finish the span
		tracing.WithTracer(tracer).
			WithOptions(tracing.SpanHttpStatusCode(gc.Writer.Status())).
			Finish(ctx)
		gc.Request = gc.Request.WithContext(orig)
	}
}
