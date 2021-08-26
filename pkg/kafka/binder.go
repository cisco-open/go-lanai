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
	properties   *KafkaProperties
	brokers      []string
	initOnce     sync.Once
	defaults     configDefaults
	globalClient sarama.Client
	adminClient  sarama.ClusterAdmin
	registry     map[string]io.Closer
	subscribers  map[string]io.Closer
	interceptors []ProducerInterceptor
}

type factoryDI struct {
	fx.In
	Lifecycle    fx.Lifecycle
	Properties   KafkaProperties
	Interceptors []ProducerInterceptor `group:"kafka"`
}

func NewSaramaProducerFactory(di factoryDI) Binder {
	s := &SaramaKafkaBinder{
		properties:  &di.Properties,
		brokers:     di.Properties.Brokers,
		registry:    make(map[string]io.Closer),
		subscribers: make(map[string]io.Closer),
		interceptors: append(di.Interceptors,
			mimeTypeProducerInterceptor{},
		),
	}
	return s
}

func (s *SaramaKafkaBinder) NewProducerWithTopic(topic string, options ...ProducerOptions) (Producer, error) {
	if _, ok := s.registry[topic]; ok {
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

	if e := s.provisionTopic(topic, &cfg); e != nil {
		return nil, e
	}

	p, err := newSaramaProducer(topic, s.brokers, &cfg)

	if err != nil {
		return nil, err
	} else {
		s.registry[topic] = p
		return p, nil
	}
}

// TODO review this
func (s *SaramaKafkaBinder) Subscribe(topic string, options ...ConsumerOptions) (*saramaSubscriber, error) {
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

func (s *SaramaKafkaBinder) ListTopics() (topics []string) {
	topics = make([]string, 0, len(s.registry))
	for t := range s.registry {
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
		prodCfg := defaultProducerConfig(cfg)
		prodCfg.interceptors = append(prodCfg.interceptors, s.interceptors...)

		// TODO consumer
		consConfig := defaultConsumerConfig(cfg)

		s.defaults = configDefaults{
			producerConfig: *prodCfg,
			consumerConfig: *consConfig,
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
	})

	return
}

// Shutdown implements BinderLifecycle, close resources
func (s *SaramaKafkaBinder) Shutdown(_ context.Context) error {
	for _, p := range s.registry {
		if e := p.Close(); e != nil {
			// since application is shutting down, we just log the error
			log.Errorf("error while closing kafka producer: %v", e)
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

func (s *SaramaKafkaBinder) provisionTopic(topic string, cfg *producerConfig) error {

	// TODO, this won't work
	//parts, e := s.globalClient.Partitions(topic)
	//fmt.Printf("%v, %v\n", parts, e)

	topicDetails := &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: 1,
		ReplicaAssignment: nil,
		ConfigEntries:     nil,
	}
	if e := s.adminClient.CreateTopic(topic, topicDetails, false); e != nil {
		return nil
	}


	return nil
}
