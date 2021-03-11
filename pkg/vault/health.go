package vault

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
)

type VaultHealthIndicator struct {
	Client *Client
}

func (i *VaultHealthIndicator) Name() string {
	return "vault"
}

func (i *VaultHealthIndicator) Health(c context.Context, options health.Options) health.Health {
	if _, e := i.Client.Sys(c).Health(); e != nil {
		return health.NewDetailedHealth(health.StatusDown, "vault /v1/sys/health failed", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "vault /v1/sys/health succeeded", nil)
	}
}

