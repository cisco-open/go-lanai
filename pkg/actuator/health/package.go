package health

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "actuator-health",
	Precedence: bootstrap.ActuatorPrecedence,
	Options: []fx.Option{
		fx.Provide(
			BindHealthProperties,
			NewSystemHealthRegistrar,
			provideInterfaces,
		),
	},
}

func Use() {
	bootstrap.Register(Module)
}

func provideInterfaces(reg *SystemHealthRegistrar) (Registrar, Indicator) {
	return reg, reg.Indicator
}

