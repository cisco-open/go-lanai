package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/prometheus/common/log"
	"go.uber.org/fx"
	"io"
	"sync"
)

type configDefaults struct {
	producerConfig
	consumerConfig
}

type SaramaKafkaBinder struct {
	properties           *KafkaProperties
	brokers              []string
	initOnce             sync.Once
	defaults             configDefaults
	producerInterceptors []ProducerMessageInterceptor
	consumerInterceptors []ConsumerDispatchInterceptor
	handlerInterceptors  []ConsumerHandlerInterceptor

	globalClient   sarama.Client
	adminClient    sarama.ClusterAdmin
	provisioner    *saramaTopicProvisioner
	producers      map[string]io.Closer
	subscribers    map[string]io.Closer
	consumerGroups map[string]io.Closer
}

type factoryDI struct {
	fx.In
	Lifecycle            fx.Lifecycle
	Properties           KafkaProperties
	ProducerInterceptors []ProducerMessageInterceptor  `group:"kafka"`
	ConsumerInterceptors []ConsumerDispatchInterceptor `group:"kafka"`
	HandlerInterceptors  []ConsumerHandlerInterceptor  `group:"kafka"`
}

func NewSaramaProducerFactory(di factoryDI) Binder {
	s := &SaramaKafkaBinder{
		properties:     &di.Properties,
		brokers:        di.Properties.Brokers,
		producers:      make(map[string]io.Closer),
		subscribers:    make(map[string]io.Closer),
		consumerGroups: make(map[string]io.Closer),
		producerInterceptors: append(di.ProducerInterceptors,
			mimeTypeProducerInterceptor{},
		),
		consumerInterceptors: di.ConsumerInterceptors,
		handlerInterceptors:  di.HandlerInterceptors,
	}
	return s
}

func (s *SaramaKafkaBinder) NewProducerWithTopic(topic string, options ...ProducerOptions) (Producer, error) {
	if _, ok := s.producers[topic]; ok {
		logger.Warnf("producer for topic %s already exist. please use the existing instance", topic)
		return nil, NewKafkaError(ErrorCodeProducerExists, "producer for topic %s already exist", topic)
	}

	if e := s.Initialize(context.Background()); e != nil {
		return nil, e
	}

	cfg := s.defaults.producerConfig
	for _, optionFunc := range options {
		optionFunc(&cfg)
	}

	if e := s.provisioner.provisionTopic(topic, &cfg); e != nil {
		return nil, e
	}

	p, err := newSaramaProducer(topic, s.brokers, &cfg)

	if err != nil {
		return nil, err
	} else {
		s.producers[topic] = p
		return p, nil
	}
}

// TODO review this
func (s *SaramaKafkaBinder) Subscribe(topic string, options ...ConsumerOptions) (Subscriber, error) {
	if _, ok := s.subscribers[topic]; ok {
		logger.Warnf("subscriber for topic %s already exist. please use the existing instance", topic)
		return nil, NewKafkaError(ErrorCodeConsumerExists, "producer for topic %s already exist", topic)
	}

	if e := s.Initialize(context.Background()); e != nil {
		return nil, e
	}

	cfg := s.defaults.consumerConfig
	for _, optionFunc := range options {
		optionFunc(&cfg)
	}

	sub, err := newSaramaSubscriber(topic, s.brokers, &cfg)
	if err != nil {
		return nil, err
	}

	s.subscribers[topic] = sub
	return sub, nil
}

func (s *SaramaKafkaBinder) Consume(topic string, group string, options ...ConsumerOptions) (GroupConsumer, error) {
	if _, ok := s.consumerGroups[topic]; ok {
		logger.Warnf("consumer group for topic %s already exist. please use the existing instance", topic)
		return nil, NewKafkaError(ErrorCodeConsumerExists, "producer for topic %s already exist", topic)
	}

	if e := s.Initialize(context.Background()); e != nil {
		return nil, e
	}

	cfg := s.defaults.consumerConfig
	for _, optionFunc := range options {
		optionFunc(&cfg)
	}

	cg, err := newSaramaGroupConsumer(topic, group, s.brokers, &cfg)
	if err != nil {
		return nil, err
	}

	s.consumerGroups[topic] = cg
	return cg, nil
}

func (s *SaramaKafkaBinder) ListTopics() (topics []string) {
	topics = make([]string, 0, len(s.producers))
	for t := range s.producers {
		topics = append(topics, t)
	}
	return topics
}

func (s *SaramaKafkaBinder) Client() sarama.Client {
	return s.globalClient
}

// Initialize implements BinderLifecycle, prepare for use, negotiate default configs, etc.
func (s *SaramaKafkaBinder) Initialize(_ context.Context) (err error) {
	s.initOnce.Do(func() {
		cfg := defaultSaramaConfig(s.properties)

		// prepare defaults
		producerCfg := defaultProducerConfig(cfg)
		producerCfg.interceptors = append(producerCfg.interceptors, s.producerInterceptors...)

		consumerCfg := defaultConsumerConfig(cfg)
		consumerCfg.dispatchInterceptors = append(consumerCfg.dispatchInterceptors, s.consumerInterceptors...)
		consumerCfg.handlerInterceptors = append(consumerCfg.handlerInterceptors, s.handlerInterceptors...)

		s.defaults = configDefaults{
			producerConfig: *producerCfg,
			consumerConfig: *consumerCfg,
		}

		// create a global client
		s.globalClient, err = sarama.NewClient(s.brokers, cfg)
		if err != nil {
			err = NewKafkaError(ErrorCodeBrokerNotReachable, fmt.Sprintf("unable to connect to Kafka brokers %v: %v", s.brokers, err), err)
			return
		}

		s.adminClient, err = sarama.NewClusterAdmin(s.brokers, cfg)
		if err != nil {
			err = NewKafkaError(ErrorCodeBrokerNotReachable, fmt.Sprintf("unable to connect to Kafka brokers %v: %v", s.brokers, err), err)
			return
		}

		s.provisioner = &saramaTopicProvisioner{
			globalClient: s.globalClient,
			adminClient:  s.adminClient,
		}
	})

	return
}

// Shutdown implements BinderLifecycle, close resources
func (s *SaramaKafkaBinder) Shutdown(_ context.Context) error {
	for _, p := range s.producers {
		if e := p.Close(); e != nil {
			// since application is shutting down, we just log the error
			log.Errorf("error while closing kafka producer: %v", e)
		}
	}

	for _, p := range s.subscribers {
		if e := p.Close(); e != nil {
			// since application is shutting down, we just log the error
			log.Errorf("error while closing kafka subscriber: %v", e)
		}
	}

	for _, p := range s.consumerGroups {
		if e := p.Close(); e != nil {
			// since application is shutting down, we just log the error
			log.Errorf("error while closing kafka consumer: %v", e)
		}
	}

	if e := s.adminClient.Close(); e != nil {
		log.Errorf("error while closing kafka admin client: %v", e)
	}

	if e := s.globalClient.Close(); e != nil {
		log.Errorf("error while closing kafka global client: %v", e)
	}
	return nil
}
