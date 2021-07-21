package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/common/log"
	"go.uber.org/fx"
	"io"
	"strings"
)

type SaramaProducerFactory struct {
	properties KafkaProperties
	registry map[string]io.Closer
}

func NewSaramaProducerFactory(lc fx.Lifecycle, p KafkaProperties) ProducerFactory {
	s := &SaramaProducerFactory{
		properties: p,
		registry: make(map[string]io.Closer),
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			for _, p := range s.registry {
				err := p.Close()
				//since application is shutting down, we just log the error
				if err != nil {
					log.Error("error while closing kafka producer %s", err)
				}
			}
			return nil
		},
	})
	return s
}

func (s *SaramaProducerFactory) NewProducerWithTopic(topic string, options ...ProducerOptions) (Producer, error) {
	if _, ok := s.registry[topic]; ok {
		logger.Warnf("producer for topic %s already exist. please use the existing instance", topic)
		return nil, errors.New(fmt.Sprintf("producer for topic %s already exist", topic))
	}

	producerConfig := defaultProducerConfig(s.properties)
	for _, optionFunc := range options {
		optionFunc(producerConfig)
	}

	brokerList := strings.Split(s.properties.Brokers, ",")

	p, err := NewSaramaProducer(topic, brokerList, producerConfig.Config)

	if err != nil {
		return nil, err
	} else {
		s.registry[topic] = p
		return p, nil
	}
}

func (s *SaramaProducerFactory) ListTopics() (topics []string) {
	for t, _ := range s.registry {
		topics = append(topics, t)
	}
	return topics
}