package tracing

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	kitopentracing "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/opentracing/opentracing-go"
)

type TracingWebCustomizer struct {
	tracer opentracing.Tracer
}

func newTracingWebCustomizer(tracer opentracing.Tracer) *TracingWebCustomizer{
	return &TracingWebCustomizer{
		tracer: tracer,
	}
}

func (c *TracingWebCustomizer) Customize(r *web.Registrar) error {
	t := kithttp.ServerBefore(kitopentracing.HTTPToContext(c.tracer,"http", logger))
	r.AddOption(t)
	return nil
}

