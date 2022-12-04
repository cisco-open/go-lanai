package opensearch

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
)

type HealthIndicator struct {
	client OpenClient
}

func (i *HealthIndicator) Name() string {
	return "opensearch"
}

func NewHealthIndicator(client OpenClient) *HealthIndicator {
	return &HealthIndicator{
		client: client,
	}
}

func (i *HealthIndicator) Health(c context.Context, options health.Options) health.Health {
	resp, err := i.client.Ping(c)
	if err != nil {
		logger.WithContext(c).Errorf("unable to ping opensearch: %v", err)
		return health.NewDetailedHealth(health.StatusDown, "opensearch ping failed", nil)
	}
	if resp.IsError() {
		logger.WithContext(c).Errorf("unable to ping opensearch: %v", resp.String())
		return health.NewDetailedHealth(health.StatusDown, "opensearch ping failed", nil)
	}
	return health.NewDetailedHealth(health.StatusUp, "opensearch ping succeeded", nil)
}
