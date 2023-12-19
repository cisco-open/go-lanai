package vaulthealth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"go.uber.org/fx"
)

type HealthRegDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	VaultClient     *vault.Client    `optional:"true"`
}

func Register(di HealthRegDI) error {
	if di.HealthRegistrar == nil || di.VaultClient == nil {
		return nil
	}
	return di.HealthRegistrar.Register(New(di.VaultClient))
}

func New(client *vault.Client) *HealthIndicator {
	return &HealthIndicator{Client: client}
}

type HealthIndicator struct {
	Client *vault.Client
}

func (i *HealthIndicator) Name() string {
	return "vault"
}

func (i *HealthIndicator) Health(c context.Context, options health.Options) health.Health {
	if _, e := i.Client.Sys(c).Health(); e != nil {
		return health.NewDetailedHealth(health.StatusDown, "vault /v1/sys/health failed", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "vault /v1/sys/health succeeded", nil)
	}
}

