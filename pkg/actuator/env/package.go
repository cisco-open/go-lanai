package env

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "actuator-env",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{
		fx.Invoke(register),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
// TODO this is not needed if we have actuator config package
func Use() {

}

type regDI struct {
	fx.In
	Registrar     *actuator.Registrar
	MgtProperties actuator.ManagementProperties
	AppContext    *bootstrap.ApplicationContext
}

func register(di regDI) {
	ep := New(di)
	di.Registrar.Register(ep)
}
