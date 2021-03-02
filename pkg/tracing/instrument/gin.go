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
		ctx := reqFunc(gc.Request.Context(), gc.Request)
		gc.Request = gc.Request.WithContext(ctx)

		gc.Next()

		// finish the span
		gc.Request = gc.Request.WithContext(
			tracing.WithTracer(tracer).
				WithOptions(tracing.SpanHttpStatusCode(gc.Writer.Status())).
				FinishAndRewind(ctx),
		)
	}
}
