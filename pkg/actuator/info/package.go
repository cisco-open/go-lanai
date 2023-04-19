package info

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("ACTR.Info")

var Module = &bootstrap.Module{
	Name: "actuator-info",
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
	AppContext    *bootstrap.ApplicationContext
}

func register(di regDI) {
	ep := newEndpoint(di)
	di.Registrar.MustRegister(ep)
}
