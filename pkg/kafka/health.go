package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"github.com/Shopify/sarama"
	"go.uber.org/fx"
	"strings"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	ProducerFactory ProducerFactory
	Properties      KafkaProperties
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil {
		return
	}

	brokerList := strings.Split(di.Properties.Brokers, ",")
	config := defaultProducerConfig(di.Properties)
	client, err := sarama.NewClient(brokerList, config.Config)

	if err != nil {
		panic(err)
	}

	di.HealthRegistrar.Register(&HealthIndicator{
		client: client,
		producerFactory: di.ProducerFactory,
	})
}

type HealthIndicator struct {
	client sarama.Client
	producerFactory ProducerFactory
}

func (i *HealthIndicator) Name() string {
	return "kafka"
}

func (i *HealthIndicator) Health(c context.Context, options health.Options) health.Health {
	topics := i.producerFactory.(*SaramaProducerFactory).ListTopics()

	err := i.client.RefreshMetadata(topics...)

	if err != nil {
		return health.NewDetailedHealth(health.StatusDown, "kafka refresh metadata failed", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "kafka refresh metadata succeeded", nil)
	}
}
