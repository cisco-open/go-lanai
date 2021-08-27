package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/loop"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/prometheus/common/log"
	"go.uber.org/fx"
	"sync"
	"time"
)

type configDefaults struct {
	producerConfig
	consumerConfig
}

type SaramaKafkaBinder struct {
	properties           *KafkaProperties
	brokers              []string
	initOnce             sync.Once
	startOnce            sync.Once
	defaults             configDefaults
	producerInterceptors []ProducerMessageInterceptor
	consumerInterceptors []ConsumerDispatchInterceptor
	handlerInterceptors  []ConsumerHandlerInterceptor
	monitor              *loop.Loop

	globalClient      sarama.Client
	adminClient       sarama.ClusterAdmin
	provisioner       *saramaTopicProvisioner
	producers         map[string]BindingLifecycle
	subscribers       map[string]BindingLifecycle
	consumerGroups    map[string]BindingLifecycle
	monitorCancelFunc context.CancelFunc
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
		properties: &di.Properties,
		brokers:    di.Properties.Brokers,
		producerInterceptors: append(di.ProducerInterceptors,
			mimeTypeProducerInterceptor{},
		),
		consumerInterceptors: di.ConsumerInterceptors,
		handlerInterceptors:  di.HandlerInterceptors,
		monitor:              loop.NewLoop(),
		producers:            make(map[string]BindingLifecycle),
		subscribers:          make(map[string]BindingLifecycle),
		consumerGroups:       make(map[string]BindingLifecycle),
	}

	if e := s.Initialize(context.Background()); e != nil {
		panic(e)
	}
	return s
}

func (s *SaramaKafkaBinder) Produce(topic string, options ...ProducerOptions) (Producer, error) {
	if _, ok := s.producers[topic]; ok {
		logger.Warnf("producer for topic %s already exist. please use the existing instance", topic)
		return nil, NewKafkaError(ErrorCodeProducerExists, "producer for topic %s already exist", topic)
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

func (s *SaramaKafkaBinder) Subscribe(topic string, options ...ConsumerOptions) (Subscriber, error) {
	if _, ok := s.subscribers[topic]; ok {
		logger.Warnf("subscriber for topic %s already exist. please use the existing instance", topic)
		return nil, NewKafkaError(ErrorCodeConsumerExists, "producer for topic %s already exist", topic)
	}

	cfg := s.defaults.consumerConfig
	for _, optionFunc := range options {
		optionFunc(&cfg)
	}

	sub, err := newSaramaSubscriber(topic, s.brokers, &cfg, s.provisioner)
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

	cfg := s.defaults.consumerConfig
	for _, optionFunc := range options {
		optionFunc(&cfg)
	}

	cg, err := newSaramaGroupConsumer(topic, group, s.brokers, &cfg, s.provisioner)
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

// Start implements BinderLifecycle, start all bindings if not started yet (Producer, Subscriber, GroupConsumer, etc).
func (s *SaramaKafkaBinder) Start(_ context.Context) (err error) {
	s.startOnce.Do(func() {
		var loopCtx context.Context
		loopCtx, s.monitorCancelFunc = s.monitor.Run(context.Background())
		s.monitor.Repeat(s.tryStartTaskFunc(loopCtx), func(opt *loop.TaskOption) {
			opt.RepeatIntervalFunc = s.tryStartRepeatIntervalFunc()
		})
	})
	return nil
}

// Shutdown implements BinderLifecycle, close resources
func (s *SaramaKafkaBinder) Shutdown(_ context.Context) error {
	if s.monitorCancelFunc != nil {
		s.monitorCancelFunc()
	}

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

// tryStartTaskFunc try to start any registered bindings if it's not started yet
// tryStartTaskFunc should be run periodically to perform delayed start of any Subscriber or GroupConsumer
func (s *SaramaKafkaBinder) tryStartTaskFunc(loopCtx context.Context) loop.TaskFunc {
	return func(_ context.Context, l *loop.Loop) (ret interface{}, err error) {
		// we cannot use passed-in context, because this context will be cancelled as soon as this function finishes
		allStarted := true
		for _, lc := range s.producers {
			if e := lc.Start(loopCtx); e != nil {
				allStarted = false
			}
		}

		for _, lc := range s.subscribers {
			if e := lc.Start(loopCtx); e != nil {
				allStarted = false
			}
		}

		for _, lc := range s.consumerGroups {
			if e := lc.Start(loopCtx); e != nil {
				allStarted = false
			}
		}

		return allStarted, nil
	}
}

// tryStartRepeatIntervalFunc decide repeat rate of tryStartTaskFunc
// we try start bindings more frequently at beginning.
// when all bindings are successfully started, we reduce the repeating rate
func (s *SaramaKafkaBinder) tryStartRepeatIntervalFunc() loop.RepeatIntervalFunc {
	return func(result interface{}, err error) time.Duration {
		switch allStarted := result.(type) {
		case bool:
			if allStarted {
				return 120 * time.Second
			} else {
				return 5 * time.Second
			}
		default:
			return 5 * time.Second
		}
	}
}
