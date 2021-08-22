package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
)

type HealthIndicator struct {
	binder SaramaBinder
}

func (i *HealthIndicator) Name() string {
	return "kafka"
}

func (i *HealthIndicator) Health(_ context.Context, opts health.Options) health.Health {
	topics := i.binder.ListTopics()

	client := i.binder.Client()
	if client == nil {
		return health.NewDetailedHealth(health.StatusUnkown, "kafka client not initialized yet", nil)
	}

	var details map[string]interface{}
	if opts.ShowDetails {
		details = map[string]interface{}{
			"topics": topics,
		}
	}

	if err := client.RefreshMetadata(topics...); err != nil {
		return health.NewDetailedHealth(health.StatusDown, "kafka refresh metadata failed", details)
	}
	return health.NewDetailedHealth(health.StatusUp, "kafka refresh metadata succeeded", details)
}
