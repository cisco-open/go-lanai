// Package healthep
// Contains implementation of health endpoint as a separate package to avoid cyclic package dependency.
//
// Implementations in this package cannot be moved to package "actuator/health", otherwise, it could create
// cyclic package dependency as following:
// 		actuator/health -> actuator -> security -> tenancy -> redis -> actuator/health
//
// Therefore, any implementations involves package mentioned above should be moved here
package healthep

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
	HealthRegistrar health.Registrar
	Registrar       *actuator.Registrar           `optional:"true"`
	MgtProperties   actuator.ManagementProperties `optional:"true"`
}

func Register(di regDI) {
	// Note: when actuator.Registrar is nil, we don't need to anything
	if di.Registrar == nil {
		return
	}
	healthReg := di.HealthRegistrar.(*health.SystemHealthRegistrar)
	endpoint, e := newEndpoint(func(opt *EndpointOption) {
		opt.MgtProperties = di.MgtProperties
		opt.Contributor = healthReg.Indicator
		opt.Properties = di.Properties
		opt.DetailsControl = healthReg.DetailsDisclosure
		opt.ComponentsControl = healthReg.ComponentsDisclosure
	})
	if e != nil {
		panic(e)
	}

	di.Registrar.MustRegister(endpoint)
}

