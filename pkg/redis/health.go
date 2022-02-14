package redis

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"go.uber.org/fx"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	RedisClient     Client
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&RedisHealthIndicator{
		client: di.RedisClient,
	})
}

type RedisHealthIndicator struct {
	client Client
}

func (i *RedisHealthIndicator) Name() string {
	return "redis"
}

func (i *RedisHealthIndicator) Health(c context.Context, options health.Options) health.Health {
	if _, e := i.client.Ping(c).Result(); e != nil {
		return health.NewDetailedHealth(health.StatusDown, "redis ping failed", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "redis ping succeeded", nil)
	}
}

