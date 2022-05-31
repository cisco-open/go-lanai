package health

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"go.uber.org/fx"
)

func init() {
	health.Use()
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

	endpoint := newEndpoint(func(opt *EndpointOption) {
		opt.MgtProperties = &di.MgtProperties
		opt.Contributor = di.HealthIndicator
		opt.Properties = &di.Properties
	})
	di.Registrar.MustRegister(endpoint)
}

