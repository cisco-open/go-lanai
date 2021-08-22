package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/prometheus/common/log"
	"go.uber.org/fx"
	"io"
	"sync"
)

type configDefaults struct {
	producerConfig
}

type SaramaProducerFactory struct {
	properties      *KafkaProperties
	brokers         []string
	initOnce        sync.Once
	defaults        configDefaults
	globalClient    sarama.Client
	registry        map[string]io.Closer
	interceptors    []ProducerInterceptor
}

type factoryDI struct {
	fx.In
	Lifecycle       fx.Lifecycle
	Properties      KafkaProperties
	Interceptors    []ProducerInterceptor `group:"kafka"`
}

func NewSaramaProducerFactory(di factoryDI) Binder {
	s := &SaramaProducerFactory{
		properties: &di.Properties,
		brokers:    di.Properties.Brokers,
		registry:   make(map[string]io.Closer),
		interceptors: append(di.Interceptors,
			mimeTypeProducerInterceptor{},
		),
	}
	return s
}

func (s *SaramaProducerFactory) NewProducerWithTopic(topic string, options ...ProducerOptions) (Producer, error) {
	if _, ok := s.registry[topic]; ok {
		logger.Warnf("producer for topic %s already exist. please use the existing instance", topic)
		return nil, errors.New(fmt.Sprintf("producer for topic %s already exist", topic))
	}

	if e := s.Initialize(context.Background()); e != nil {
		return nil, e
	}

	cfg := s.defaults.producerConfig
	for _, optionFunc := range options {
		optionFunc(&cfg)
	}

	p, err := newSaramaProducer(topic, s.brokers, &cfg)

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

func (s *SaramaProducerFactory) Client() sarama.Client {
	return s.globalClient
}

// Initialize implements BinderLifecycle, prepare for use, negotiate default configs, etc.
func (s *SaramaProducerFactory) Initialize(_ context.Context) (err error) {
	s.initOnce.Do(func() {
		cfg := defaultSaramaConfig(s.properties)

		// prepare defaults
		prodCfg := defaultProducerConfig(cfg)
		prodCfg.interceptors = append(prodCfg.interceptors, s.interceptors...)

		// TODO consumer

		s.defaults = configDefaults{
			producerConfig: *prodCfg,
		}

		// create a global client
		s.globalClient, err = sarama.NewClient(s.brokers, cfg)
		if err != nil {
			err = fmt.Errorf("unable to connect to Kafka brokers %v: %v", s.brokers, err)
			return
		}
	})

	return
}

// Shutdown implements BinderLifecycle, close resources
func (s *SaramaProducerFactory) Shutdown(_ context.Context) error {
	for _, p := range s.registry {
		if e := p.Close(); e != nil {
			// since application is shutting down, we just log the error
			log.Errorf("error while closing kafka producer %s", e)
		}
	}

	if e := s.globalClient.Close(); e != nil {
		log.Errorf("error while closing kafka producer %s", e)
	}
	return nil
}
