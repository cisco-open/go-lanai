package health

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "actuator-health",
	Precedence: bootstrap.ActuatorPrecedence,
	Options: []fx.Option{
		fx.Provide(provide, BindHealthProperties),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type provideDI struct {
	fx.In
	Properties    HealthProperties
}

type provideOut struct {
	fx.Out
	Registrar       Registrar
	HealthIndicator *SystemHealthIndicator
}

// provide SystemHealthIndicator as both Registrar and *SystemHealthIndicator
func provide(di provideDI) (out provideOut) {
	indicator := NewSystemHealthIndicator(di)
	return provideOut{
		Registrar: indicator,
		HealthIndicator: indicator,
	}
}
