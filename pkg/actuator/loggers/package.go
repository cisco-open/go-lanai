package loggers

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

//var logger = log.New("ACTR.LoggerLevel")

var Module = &bootstrap.Module{
	Name:       "actuator-loggers",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func Register() {
	bootstrap.Register(Module)
}

type regDI struct {
	fx.In
	Registrar     *actuator.Registrar
	MgtProperties actuator.ManagementProperties
}

func register(di regDI) {
	ep := newEndpoint(di)
	di.Registrar.MustRegister(ep)
}
