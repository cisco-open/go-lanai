package webtracing

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "web-tracing",
	Precedence: web.MinWebPrecedence,
	PriorityOptions: []fx.Option{
		fx.Invoke(setup),
	},
}

type initDI struct {
	fx.In
	Registrar *web.Registrar `optional:"true"`
	Tracer    opentracing.Tracer `optional:"true"`
}

func setup(di initDI) {
	if di.Tracer != nil && di.Registrar != nil {
		di.Registrar.MustRegister(newTracingWebCustomizer(di.Tracer))
	}
}
