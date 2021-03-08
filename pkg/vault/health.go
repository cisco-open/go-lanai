package vault

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"go.uber.org/fx"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	VaultClient     *Client
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil {
		return
	}
	di.HealthRegistrar.Register(&VaultHealthIndicator{
		client: di.VaultClient,
	})
}

type VaultHealthIndicator struct {
	client *Client
}

func (i *VaultHealthIndicator) Name() string {
	return "redis"
}

func (i *VaultHealthIndicator) Health(c context.Context, options health.Options) health.Health {
	if _, e := i.client.Sys().Health(); e != nil {
		return health.NewDetailedHealth(health.StatusDown, "vault /v1/sys/health failed", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "vault /v1/sys/health succeeded", nil)
	}
}

