package actuator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("Actuator")

var Module = &bootstrap.Module{
	Name: "actuate",
	Precedence: MaxActuatorPrecedence,
	Options: []fx.Option{
		fx.Provide(NewRegistrar, BindManagementProperties),
		fx.Invoke(initialize),
	},
}

func init() {
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

/**************************
	Initialize
***************************/
func initialize(registrar *Registrar, di initDI) {
	if e := registrar.initialize(di); e != nil {
		panic(e)
	}
}


