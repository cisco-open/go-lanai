package health

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "actuator-health",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{
		fx.Provide(provide, BindHealthProperties),
	},
}

func init() {
	bootstrap.Register(Module)
}

type provideDI struct {
	fx.In
	MgtProperties actuator.ManagementProperties `optional:"true"`
	Properties    HealthProperties
}

type provideOut struct {
	fx.Out
	Registrar Registrar
	HealthIndicator *SystemHealthIndicator
}

// provide SystemHealthIndicator as both Registrar and *SystemHealthIndicator
func provide(di provideDI) (out provideOut) {
	indicator := newSystemHealthIndicator(di)
	return provideOut{
		Registrar: indicator,
		HealthIndicator: indicator,
	}
}

type regDI struct {
	fx.In
	Properties      HealthProperties
	Registrar       *actuator.Registrar           `optional:"true"`
	MgtProperties   actuator.ManagementProperties `optional:"true"`
	HealthIndicator *SystemHealthIndicator        `optional:"true"`
}

func Register(di regDI) {
	// Note: when actuator.Registrar is nil, we don't need to anything
	if di.Registrar == nil {
		return
	}
	initialize(di)
}
