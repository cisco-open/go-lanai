package opainit

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"go.uber.org/fx"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	OPAReady        opa.EmbeddedOPAReadyCH
}

func RegisterHealth(di regDI) {
	if di.HealthRegistrar == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&HealthIndicator{
		ready: di.OPAReady,
	})
}

type HealthIndicator struct {
	ready opa.EmbeddedOPAReadyCH
}

func (i *HealthIndicator) Name() string {
	return "opa"
}

func (i *HealthIndicator) Health(_ context.Context, _ health.Options) health.Health {
	select {
	case <-i.ready:
		return health.NewDetailedHealth(health.StatusUp, "OPA engine is UP", nil)
	default:
		return health.NewDetailedHealth(health.StatusDown, "OPA engine is not ready", nil)
	}
}
