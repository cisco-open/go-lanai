package kafka

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	"go.uber.org/fx"
)

var logger = log.New("Kafka")

var Module = &bootstrap.Module{
	Precedence: bootstrap.KafkaPrecedence,
	Options: []fx.Option{
		fx.Provide(BindKafkaProperties),
		fx.Provide(NewKafkaBinder),
		fx.Invoke(initialize),
	},
}

const (
	FxGroup = "kafka"
)

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(tlsconfig.Module)
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	Lifecycle       fx.Lifecycle
	Properties      KafkaProperties
	Binder          Binder
	HealthRegistrar health.Registrar `optional:"true"`
}

func initialize(di initDI) {

	// register lifecycle functions
	di.Lifecycle.Append(fx.Hook{
		OnStart: di.Binder.(BinderLifecycle).Start,
		OnStop:  di.Binder.(BinderLifecycle).Shutdown,
	})

	// register health endpoints if applicable
	if di.HealthRegistrar == nil {
		return
	}

	di.HealthRegistrar.MustRegister(&HealthIndicator{
		binder: di.Binder.(SaramaBinder),
	})
}
