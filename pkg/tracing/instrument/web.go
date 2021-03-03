package instrument

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
)

type TracingWebCustomizer struct {
	tracer opentracing.Tracer
}

func NewTracingWebCustomizer(tracer opentracing.Tracer) *TracingWebCustomizer{
	return &TracingWebCustomizer{
		tracer: tracer,
	}
}

// we want TracingWebCustomizer before anything else
func (c TracingWebCustomizer) Order() int {
	return order.Highest
}

func (c *TracingWebCustomizer) Customize(ctx context.Context, r *web.Registrar) error {
	// for gin
	r.AddGlobalMiddlewares(GinTracing(c.tracer, tracing.OpNameHttp))

	// for go-kit endpoints, because we are unable to finish the created span,
	// so we rely on Gin middleware to create/finish span
	//t := kithttp.ServerBefore(kitopentracing.HTTPToContext(c.tracer, tracing.OpNameHttp, logger))
	//r.AddOption(t)
	return nil
}
