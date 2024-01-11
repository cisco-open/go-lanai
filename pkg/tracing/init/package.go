package tracing

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/scheduler"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing/instrument"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/fx"
)

var logger = log.New("Tracing")

var Module = &bootstrap.Module{
	Name:       "Tracing",
	Precedence: bootstrap.TracingPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(tracing.BindTracingProperties),
		fx.Provide(provideTracer),
		fx.Provide(instrument.GormTracingProvider()),
		fx.Provide(instrument.CliRunnerTracingProvider()),
		fx.Provide(instrument.HttpClientTracingProvider()),
		fx.Provide(instrument.SecurityScopeTracingProvider()),
		fx.Provide(instrument.KafkaTracingTracingProvider()),
		fx.Provide(instrument.OpenSearchTracingProvider()),
		fx.Invoke(initialize),
	},
}

type TracerClosingHook *fx.Hook

var defaultTracerCloser fx.Hook

type kCtxDefaultTracerCloser struct {}

// Use does nothing. Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
	EnableBootstrapTracing(bootstrap.GlobalBootstrapper())
}

// EnableBootstrapTracing enable bootstrap tracing on a given bootstrapper.
// bootstrap.GlobalBootstrapper() should be used for regular application that uses bootstrap.Execute()
func EnableBootstrapTracing(bootstrapper *bootstrap.Bootstrapper) {
	// logger extractor
	log.RegisterContextLogFields(tracing.TracingLogValuers)

	appTracer, closer := tracing.NewDefaultTracer()
	bootstrapper.AddInitialAppContextOptions(instrument.MakeBootstrapTracingOption(appTracer, tracing.OpNameBootstrap))
	bootstrapper.AddStartContextOptions(instrument.MakeStartTracingOption(appTracer, tracing.OpNameStart))
	bootstrapper.AddStopContextOptions(instrument.MakeStopTracingOption(appTracer, tracing.OpNameStop))
	defaultTracerCloser = fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.WithContext(ctx).Infof("closing default Tracer...")
			e := closer.Close()
			if e != nil {
				logger.WithContext(ctx).Errorf("failed to close default Tracer: %v", e)
			}
			logger.WithContext(ctx).Infof("default Tracer closed")
			return e
		},
	}
	bootstrapper.AddInitialAppContextOptions(func(ctx context.Context) context.Context {
		return context.WithValue(ctx, kCtxDefaultTracerCloser{}, defaultTracerCloser)
	})
}

/**************************
	Provide dependencies
***************************/
type tracerOut struct {
	fx.Out
	Tracer opentracing.Tracer
	FxHook TracerClosingHook
}

func provideTracer(ctx *bootstrap.ApplicationContext, props tracing.TracingProperties) (ret tracerOut) {
	ret = tracerOut{
		Tracer: opentracing.NoopTracer{},
	}

	if !props.Enabled {
		return
	}

	tracers := make([]opentracing.Tracer, 0, 2)
	if props.Jaeger.Enabled {
		tracer, closer := tracing.NewJaegerTracer(ctx, &props.Jaeger, &props.Sampler)
		tracers = append(tracers, tracer)
		ret.FxHook = &fx.Hook{
			OnStop: func(ctx context.Context) error {
				logger.WithContext(ctx).Infof("closing Jaeger Tracer...")
				e := closer.Close()
				if e != nil {
					logger.WithContext(ctx).Errorf("failed to close Jaeger Tracer: %v", e)
				}
				logger.WithContext(ctx).Infof("Jaeger Tracer closed")
				return e
			},
		}
	}

	if props.Zipkin.Enabled {
		panic("zipkin is currently unsupported")
	}

	switch len(tracers) {
	case 0:
		return
	case 1:
		ret.Tracer = tracers[0]
		return
	default:
		panic("multiple opentracing.Tracer detected. we currely only support single tracer")
	}
}

/**************************
	Setup
***************************/
type regDI struct {
	fx.In
	AppContext   *bootstrap.ApplicationContext
	Tracer       opentracing.Tracer  `optional:"true"`
	FxHook       TracerClosingHook   `optional:"true"`
	Registrar    *web.Registrar      `optional:"true"`
	RedisFactory redis.ClientFactory `optional:"true"`
	VaultClient  *vault.Client       `optional:"true"`
	// we could include security configurations, customizations here
}

func initialize(lc fx.Lifecycle, di regDI) {
	if di.Tracer == nil {
		return
	}

	// web instrumentation
	if di.Registrar != nil {
		customizer := instrument.NewTracingWebCustomizer(di.Tracer)
		if e := di.Registrar.Register(customizer); e != nil {
			panic(e)
		}
	}

	// redis instrumentation
	if di.RedisFactory != nil {
		hook := instrument.NewRedisTrackingHook(di.Tracer)
		di.RedisFactory.AddHooks(di.AppContext, hook)
	}

	// vault instrumentation
	if di.VaultClient != nil {
		hook := instrument.NewVaultTracingHook(di.Tracer)
		di.VaultClient.AddHooks(di.AppContext, hook)
	}

	// scheduler instrumentation
	scheduler.AddDefaultHook(instrument.NewTracingTaskHook(di.Tracer))

	// graceful closer
	if di.FxHook != nil {
		lc.Append(*di.FxHook)
		if defaultCloserFromCtx, ok := di.AppContext.Value(kCtxDefaultTracerCloser{}).(fx.Hook); ok {
			lc.Append(defaultCloserFromCtx)
		} else {
			lc.Append(defaultTracerCloser)
		}
	}
}
