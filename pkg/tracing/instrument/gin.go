package instrument

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"github.com/gin-gonic/gin"
	kitopentracing "github.com/go-kit/kit/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
)

func GinTracing(tracer opentracing.Tracer, opName string) gin.HandlerFunc {
	reqFunc := kitopentracing.HTTPToContext(tracer, opName, logger)
	return func(gc *gin.Context) {
		// start or join span
		orig := gc.Request.Context()
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
