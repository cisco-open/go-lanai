package loggers

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

//var logger = log.New("ACTR.LoggerLevel")

var Module = &bootstrap.Module{
	Name:       "actuator-apilist",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{},
}

func init() {
	bootstrap.Register(Module)
}

type regDI struct {
	fx.In
	Registrar     *actuator.Registrar
	MgtProperties actuator.ManagementProperties
}

func Register(di regDI) {
	ep := newEndpoint(di)
	di.Registrar.MustRegister(ep)
}
