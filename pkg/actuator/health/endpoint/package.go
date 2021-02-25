package health

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "actuator-health",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{
	},
}

func init() {
	bootstrap.Register(Module)
}

type regDI struct {
	fx.In
	Properties      health.HealthProperties
	Registrar       *actuator.Registrar           `optional:"true"`
	MgtProperties   actuator.ManagementProperties `optional:"true"`
	HealthIndicator *health.SystemHealthIndicator `optional:"true"`
}

func Register(di regDI) {
	// Note: when actuator.Registrar is nil, we don't need to anything
	if di.Registrar == nil {
		return
	}
	initialize(di)
}

func initialize(di regDI) error {

	endpoint := newEndpoint(func(opt *EndpointOption) {
		opt.MgtProperties = &di.MgtProperties
		opt.Contributor = di.HealthIndicator
		opt.Properties = &di.Properties
	})
	di.Registrar.Register(endpoint)
	return nil
}
