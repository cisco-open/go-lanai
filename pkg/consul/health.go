package consul

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
)

type ConsulHealthIndicator struct {
	conn *Connection
}

func NewConsulHealthIndicator(conn *Connection) *ConsulHealthIndicator {
	return &ConsulHealthIndicator{
		conn: conn,
	}
}

func (i *ConsulHealthIndicator) Name() string {
	return "consul"
}

func (i *ConsulHealthIndicator) Health(c context.Context, options health.Options) health.Health {

	if _, e := i.conn.Client().Status().Leader(); e != nil {
		return health.NewDetailedHealth(health.StatusDown, "consul leader status failed", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "consul leader status succeeded", nil)
	}
}
