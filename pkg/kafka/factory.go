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
	properties   *KafkaProperties
	registry     map[string]io.Closer
	interceptors []ProducerInterceptor
}

type factoryDI struct {
	fx.In
	Lifecycle    fx.Lifecycle
	Properties   KafkaProperties
	Interceptors []ProducerInterceptor `group:"kafka"`
}

func NewSaramaProducerFactory(di factoryDI) ProducerFactory {
	s := &SaramaProducerFactory{
		properties: &di.Properties,
		registry:   make(map[string]io.Closer),
		interceptors: append(di.Interceptors,
			mimeTypeProducerInterceptor{},
		),
	}

	di.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			for _, p := range s.registry {
				err := p.Close()
				//since application is shutting down, we just log the error
				if err != nil {
					log.Errorf("error while closing kafka producer %s", err)
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

	cfg := defaultProducerConfig(s.properties)
	cfg.interceptors = append(cfg.interceptors, s.interceptors...)
	for _, optionFunc := range options {
		optionFunc(cfg)
	}

	brokerList := strings.Split(s.properties.Brokers, ",")

	p, err := newSaramaProducer(topic, brokerList, cfg)

	if err != nil {
		return nil, err
	} else {
		s.registry[topic] = p
		return p, nil
	}
}

func (s *SaramaProducerFactory) ListTopics() (topics []string) {
	topics = make([]string, 0, len(s.registry))
	for t := range s.registry {
		topics = append(topics, t)
	}
	return topics
}
