package tracing

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/fx"
)

var logger = log.New("Tracing")
// for now, appTracer is used for application startup/shutdown tracing
var appTracer opentracing.Tracer

var Module = &bootstrap.Module{
	Name: "Tracing",
	Precedence: web.MinWebPrecedence + 1,
	PriorityOptions: []fx.Option{
		fx.Provide(newTracer),
		fx.Invoke(initialize),
	},
}

func init() {
	bootstrap.Register(Module)
	// bootstrap tracing
	appTracer, _ = jaeger.NewTracer("lanai", jaeger.NewConstSampler(false), jaeger.NewNullReporter())
	bootstrap.AddStartContextOptions(tracing.MakeBootstrapTracingOption(appTracer, "startup"))
	bootstrap.AddStopContextOptions(tracing.MakeBootstrapTracingOption(appTracer, "shutdown"))
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

/**************************
	Provide dependencies
***************************/
func newTracer() opentracing.Tracer {
	// TODO use Jaeger or Zipkin tracer based on properties
	// TODO properly store returned Closer and hook it up with application lifecycle
	tracer, _ := jaeger.NewTracer("lanai", jaeger.NewConstSampler(false), jaeger.NewNullReporter())
	return tracer
}

/**************************
	Setup
***************************/
type regDI struct {
	fx.In
	Registrar *web.Registrar
	Tracer    opentracing.Tracer
	// we could include security configurations, customizations here
}

func initialize(di regDI) {
	// web instrumentation
	customizer := newTracingWebCustomizer(di.Tracer)
	if e := di.Registrar.Register(customizer); e != nil {
		panic(e)
	}

	// logger extractor
	log.RegisterContextLogFields(tracingLogValuers)
}

