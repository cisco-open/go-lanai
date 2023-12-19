package consulhealth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"go.uber.org/fx"
)

type HealthIndicator struct {
	conn *consul.Connection
}

type HealthRegDI struct {
	fx.In
	HealthRegistrar health.Registrar   `optional:"true"`
	ConsulClient    *consul.Connection `optional:"true"`
}

func Register(di HealthRegDI) error {
	if di.HealthRegistrar == nil || di.ConsulClient == nil {
		return nil
	}
	return di.HealthRegistrar.Register(New(di.ConsulClient))
}

func New(conn *consul.Connection) *HealthIndicator {
	return &HealthIndicator{
		conn: conn,
	}
}

func (i *HealthIndicator) Name() string {
	return "consul"
}

func (i *HealthIndicator) Health(c context.Context, options health.Options) health.Health {

	if _, e := i.conn.Client().Status().Leader(); e != nil {
		return health.NewDetailedHealth(health.StatusDown, "consul leader status failed", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "consul leader status succeeded", nil)
	}
}
