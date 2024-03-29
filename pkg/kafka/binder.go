// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/certs"
	"github.com/cisco-open/go-lanai/pkg/utils/loop"
	"io"
	"math"
	"strings"
	"sync"
	"time"
)

const (
	errTmplProducerExists       = `producer for topic %s already exist. please use the existing instance`
	errTmplSubscriberExists     = `subscriber for topic %s already exist. please use the existing instance`
	errTmplConsumerGroupExists  = `consumer group for topic %s already exist. please use the existing instance`
	errTmplCannotConnectBrokers = `unable to connect to Kafka brokers %v: %v`
)

type SaramaKafkaBinder struct {
	sync.RWMutex
	appConfig            bootstrap.ApplicationConfig
	properties           *KafkaProperties
	brokers              []string
	initOnce             sync.Once
	startOnce            sync.Once
	defaults             bindingConfig
	producerInterceptors []ProducerMessageInterceptor
	consumerInterceptors []ConsumerDispatchInterceptor
	handlerInterceptors  []ConsumerHandlerInterceptor
	monitor              *loop.Loop
	tlsCertsManager      certs.Manager

	// TODO consider mutex lock for following fields
	producers      map[string]BindingLifecycle
	subscribers    map[string]BindingLifecycle
	consumerGroups map[string]BindingLifecycle

	// following fields are protected by mutex lock
	globalClient      sarama.Client
	adminClient       sarama.ClusterAdmin
	tlsSource         certs.Source
	provisioner       *saramaTopicProvisioner
	closed            bool
	monitorCtx        context.Context
	monitorCancelFunc context.CancelFunc
}

type BinderOptions func(opt *BinderOption)

type BinderOption struct {
	ApplicationConfig    bootstrap.ApplicationConfig
	Properties           KafkaProperties
	ProducerInterceptors []ProducerMessageInterceptor
	ConsumerInterceptors []ConsumerDispatchInterceptor
	HandlerInterceptors  []ConsumerHandlerInterceptor
	TLSCertsManager      certs.Manager
}

func NewBinder(ctx context.Context, opts ...BinderOptions) *SaramaKafkaBinder {
	opt := BinderOption{
		ProducerInterceptors: []ProducerMessageInterceptor{mimeTypeProducerInterceptor{}},
	}
	for _, fn := range opts {
		fn(&opt)
	}
	properties := opt.Properties
	s := &SaramaKafkaBinder{
		appConfig:            opt.ApplicationConfig,
		properties:           &properties,
		brokers:              opt.Properties.Brokers,
		producerInterceptors: opt.ProducerInterceptors,
		consumerInterceptors: opt.ConsumerInterceptors,
		handlerInterceptors:  opt.HandlerInterceptors,
		monitor:              loop.NewLoop(),
		producers:            make(map[string]BindingLifecycle),
		subscribers:          make(map[string]BindingLifecycle),
		consumerGroups:       make(map[string]BindingLifecycle),
		tlsCertsManager:      opt.TLSCertsManager,
	}

	if e := s.Initialize(ctx); e != nil {
		panic(e)
	}
	return s
}

func (b *SaramaKafkaBinder) prepareDefaults(ctx context.Context, saramaDefaults *sarama.Config) {
	b.defaults = bindingConfig{
		name:       "default",
		properties: BindingProperties{},
		sarama:     *saramaDefaults,
		msgLogger:  newSaramaMessageLogger(),
		producer: producerConfig{
			keyEncoder:   binaryEncoder{},
			interceptors: b.producerInterceptors,
			provisioning: topicConfig{
				autoCreateTopic:      true,
				autoAddPartitions:    true,
				allowLowerPartitions: true,
				partitionCount:       1,
				replicationFactor:    1,
			},
		},
		consumer: consumerConfig{
			dispatchInterceptors: b.consumerInterceptors,
			handlerInterceptors:  b.handlerInterceptors,
			msgLogger:            newSaramaMessageLogger(),
		},
	}

	// try load default properties
	if e := b.appConfig.Bind(&b.defaults.properties, ConfigKafkaDefaultBindingPrefix); e != nil {
		logger.WithContext(ctx).Infof("default kafka binding properties [%s.*] is not configured")
	}
}

// CloseProducer release resources for dynamic producers
func (b *SaramaKafkaBinder) CloseProducer(ctx context.Context, topic string) {
	if p, ok := b.producers[topic]; ok {
		if e := p.Close(); e != nil {
			logger.WithContext(ctx).Errorf("error while closing kafka producer: %v", e)
		}
	}
	delete(b.producers, topic)
}

func (b *SaramaKafkaBinder) Produce(topic string, options ...ProducerOptions) (Producer, error) {
	if p, ok := b.producers[topic]; ok && !p.Closed() {
		logger.Warnf(errTmplProducerExists, topic)
		return nil, NewKafkaError(ErrorCodeProducerExists, errTmplProducerExists, topic)
	}

	// apply defaults and options
	cfg := b.defaults // make a copy
	cfg.name = strings.ToLower(topic)
	for _, optionFunc := range options {
		optionFunc(&cfg)
	}

	// load and apply properties
	props := b.loadProperties(cfg.name)
	WithProducerProperties(&props.Producer)(&cfg)

	if e := b.provisioner.provisionTopic(topic, &cfg); e != nil {
		return nil, e
	}

	p, err := newSaramaProducer(topic, b.brokers, &cfg)
	if err != nil {
		return nil, err
	}

	b.producers[topic] = p
	return p, b.tryScheduleStart(p)
}

func (b *SaramaKafkaBinder) Subscribe(topic string, options ...ConsumerOptions) (Subscriber, error) {
	if s, ok := b.subscribers[topic]; ok && !s.Closed() {
		logger.Warnf(errTmplSubscriberExists, topic)
		return nil, NewKafkaError(ErrorCodeConsumerExists, errTmplSubscriberExists, topic)
	}

	// apply defaults and options
	cfg := b.defaults // make a copy
	cfg.name = strings.ToLower(topic)
	for _, optionFunc := range options {
		optionFunc(&cfg)
	}

	// load and apply properties
	props := b.loadProperties(cfg.name)
	WithConsumerProperties(&props.Consumer)(&cfg)

	sub, err := newSaramaSubscriber(topic, b.brokers, &cfg, b.provisioner)
	if err != nil {
		return nil, err
	}

	b.subscribers[topic] = sub
	return sub, b.tryScheduleStart(sub)
}

func (b *SaramaKafkaBinder) Consume(topic string, group string, options ...ConsumerOptions) (GroupConsumer, error) {
	if c, ok := b.consumerGroups[topic]; ok && !c.Closed() {
		logger.Warnf(errTmplConsumerGroupExists, topic)
		return nil, NewKafkaError(ErrorCodeConsumerExists, errTmplConsumerGroupExists, topic)
	}

	// apply defaults and options
	cfg := b.defaults // make a copy
	cfg.name = strings.ToLower(topic)
	for _, optionFunc := range options {
		optionFunc(&cfg)
	}

	// load and apply properties
	props := b.loadProperties(cfg.name)
	WithConsumerProperties(&props.Consumer)(&cfg)

	cg, err := newSaramaGroupConsumer(topic, group, b.brokers, &cfg, b.provisioner)
	if err != nil {
		return nil, err
	}

	b.consumerGroups[topic] = cg
	return cg, b.tryScheduleStart(cg)
}

func (b *SaramaKafkaBinder) ListTopics() (topics []string) {
	topics = make([]string, 0, len(b.producers)+len(b.subscribers)+len(b.consumerGroups))
	for t := range b.producers {
		topics = append(topics, t)
	}
	for t := range b.subscribers {
		topics = append(topics, t)
	}
	for t := range b.consumerGroups {
		topics = append(topics, t)
	}
	return topics
}

func (b *SaramaKafkaBinder) Client() sarama.Client {
	return b.globalClient
}

// Initialize implements BinderLifecycle, prepare for use, negotiate default configs, etc.
func (b *SaramaKafkaBinder) Initialize(ctx context.Context) (err error) {
	b.initOnce.Do(func() {
		b.Lock()
		defer b.Unlock()
		if b.closed {
			err = ErrorStartClosedBinding.WithMessage("attempt to initialize Binder after shutdown")
			return
		}
		cfg, e := defaultSaramaConfig(ctx, b.properties)
		if e != nil {
			err = NewKafkaError(ErrorCodeBindingInternal, fmt.Sprintf("unable to create kafka config: %v", e))
			logger.WithContext(ctx).Errorf("%v", err)
			return
		}

		// config TLS if enabled
		if b.properties.Net.Tls.Enable {
			if b.tlsCertsManager == nil {
				err = fmt.Errorf("failed to initialize Binder: TLS Auth is enabled but certificate manager is not provisioned")
				return
			}
			b.tlsSource, err = b.tlsCertsManager.Source(ctx, certs.WithSourceProperties(&b.properties.Net.Tls.Certs))
			if err != nil {
				logger.WithContext(ctx).Errorf("failed to get tls provider: %s", err.Error())
				return
			}
			cfg.Net.TLS.Enable = true
			cfg.Net.TLS.Config, err = b.tlsSource.TLSConfig(ctx)
			if err != nil {
				logger.WithContext(ctx).Errorf("Failed to initialize Kafka binder: %v", err)
				return
			}
		}

		// prepare defaults
		b.prepareDefaults(ctx, cfg)

		// create a global client
		b.globalClient, err = sarama.NewClient(b.brokers, cfg)
		if err != nil {
			err = NewKafkaError(ErrorCodeBrokerNotReachable, fmt.Sprintf(errTmplCannotConnectBrokers, b.brokers, err), err)
			logger.WithContext(ctx).Errorf("%v", err)
			return
		}

		b.adminClient, err = sarama.NewClusterAdmin(b.brokers, cfg)
		if err != nil {
			err = NewKafkaError(ErrorCodeBrokerNotReachable, fmt.Sprintf(errTmplCannotConnectBrokers, b.brokers, err), err)
			logger.WithContext(ctx).Errorf("%v", err)
			return
		}

		b.provisioner = &saramaTopicProvisioner{
			globalClient: b.globalClientProvider,
			adminClient:  b.clusterAdminProvider,
		}
	})

	return
}

// Start implements BinderLifecycle, start all bindings if not started yet (Producer, Subscriber, GroupConsumer, etc).
func (b *SaramaKafkaBinder) Start(ctx context.Context) (err error) {
	b.startOnce.Do(func() {
		b.Lock()
		defer b.Unlock()
		if b.closed {
			err = ErrorStartClosedBinding.WithMessage("attempt to initialize Binder after shutdown")
			return
		}
		b.monitorCtx, b.monitorCancelFunc = b.monitor.Run(ctx)
		//nolint:contextcheck // b.monitorCtx is derived from given context
		b.monitor.Repeat(b.tryStartTaskFunc(b.monitorCtx), func(opt *loop.TaskOption) {
			opt.RepeatIntervalFunc = b.tryStartRepeatIntervalFunc()
		})
		//nolint:contextcheck // b.monitorCtx is derived from given context
		go func(c context.Context) {
			select {
			case <-c.Done():
				_ = b.Shutdown(ctx)
			}
		}(b.monitorCtx)
	})
	return nil
}

// Shutdown implements BinderLifecycle, close resources
func (b *SaramaKafkaBinder) Shutdown(ctx context.Context) error {
	b.Lock()
	defer b.Unlock()
	defer func() { b.closed = true }()
	if b.monitorCancelFunc == nil {
		return nil
	}

	logger.WithContext(ctx).Infof("Kafka shutting down")
	logger.WithContext(ctx).Debugf("stopping binding watchdog...")
	b.monitorCancelFunc()
	b.monitorCtx = nil
	b.monitorCancelFunc = nil

	logger.WithContext(ctx).Debugf("closing producers...")
	for _, p := range b.producers {
		if e := p.Close(); e != nil {
			// since application is shutting down, we just log the error
			logger.WithContext(ctx).Errorf("error while closing kafka producer: %v", e)
		}
	}

	logger.WithContext(ctx).Debugf("closing subscribers...")
	for _, p := range b.subscribers {
		if e := p.Close(); e != nil {
			// since application is shutting down, we just log the error
			logger.WithContext(ctx).Errorf("error while closing kafka subscriber: %v", e)
		}
	}

	logger.WithContext(ctx).Debugf("closing group consumers...")
	for _, p := range b.consumerGroups {
		if e := p.Close(); e != nil {
			// since application is shutting down, we just log the error
			logger.WithContext(ctx).Errorf("error while closing kafka consumer: %v", e)
		}
	}

	logger.WithContext(ctx).Debugf("closing connections...")
	if e := b.adminClient.Close(); e != nil {
		logger.WithContext(ctx).Errorf("error while closing kafka admin client: %v", e)
	}

	if e := b.globalClient.Close(); e != nil {
		logger.WithContext(ctx).Errorf("error while closing kafka global client: %v", e)
	}

	if closer, ok := b.tlsSource.(io.Closer); ok {
		if e := closer.Close(); e != nil {
			logger.WithContext(ctx).Errorf("error while closing tls config provider: %v", e)
		}
	}

	logger.WithContext(ctx).Infof("Kafka connections closed")
	return nil
}

func (b *SaramaKafkaBinder) Done() <-chan struct{} {
	b.RLock()
	defer b.RUnlock()
	if b.monitorCtx != nil {
		return b.monitorCtx.Done()
	}
	// called after "Shutdown", return a closed channel
	done := make(chan struct{}, 1)
	close(done)
	return done
}

// loadProperties load properties for particular topic
func (b *SaramaKafkaBinder) loadProperties(name string) *BindingProperties {
	prefix := ConfigKafkaBindingPrefix + "." + strings.ToLower(name)
	props := b.defaults.properties // make a copy
	if e := b.appConfig.Bind(&props, prefix); e != nil {
		props = b.defaults.properties // make a fresh copy
	}
	return &props
}

func (b *SaramaKafkaBinder) globalClientProvider() (sarama.Client, error) {
	return b.globalClient, nil
}

func (b *SaramaKafkaBinder) clusterAdminProvider() (sarama.ClusterAdmin, error) {
	// simple test to see if admin client is still working
	filter := sarama.AclFilter{
		ResourceType: sarama.AclResourceTopic,
		Operation:    sarama.AclOperationRead,
	}
	_, e := b.adminClient.ListAcls(filter)
	if e == nil {
		return b.adminClient, nil
	}

	newClient, e := sarama.NewClusterAdmin(b.brokers, &b.defaults.sarama)
	if e != nil {
		return nil, NewKafkaError(ErrorCodeBrokerNotReachable, fmt.Sprintf(errTmplCannotConnectBrokers, b.brokers, e), e)
	}
	_ = b.adminClient.Close()
	b.adminClient = newClient
	return newClient, nil
}

// tryScheduleStart try to schedule start given BindingLifecycle using monitor loop if started, otherwise do nothing
func (b *SaramaKafkaBinder) tryScheduleStart(lc BindingLifecycle) error {
	b.RLock()
	defer b.RUnlock()
	if b.monitorCtx != nil {
		b.monitor.Do(b.tryStartSingleTaskFunc(b.monitorCtx, lc))
	}
	return nil
}

// tryStartSingleTaskFunc try to start given Binding
func (b *SaramaKafkaBinder) tryStartSingleTaskFunc(loopCtx context.Context, lc BindingLifecycle) loop.TaskFunc {
	return func(_ context.Context, l *loop.Loop) (ret interface{}, err error) {
		// we cannot use passed-in context, because this context will be cancelled as soon as this function finishes
		e := lc.Start(loopCtx)
		return e == nil, e
	}
}

// tryStartTaskFunc try to start any registered bindings if it's not started yet
// this task should be run periodically to perform delayed start of any Subscriber or GroupConsumer
func (b *SaramaKafkaBinder) tryStartTaskFunc(loopCtx context.Context) loop.TaskFunc {
	return func(_ context.Context, l *loop.Loop) (ret interface{}, err error) {
		// we cannot use passed-in context, because this context will be cancelled as soon as this function finishes
		allStarted := true
		toProcess := []map[string]BindingLifecycle{
			b.producers, b.subscribers, b.consumerGroups,
		}
		for _, bindings := range toProcess {
			for k, lc := range bindings {
				switch e := lc.Start(loopCtx); {
				case errors.Is(e, ErrorStartClosedBinding):
					delete(bindings, k)
				case e != nil:
					allStarted = false
				}
			}
		}
		return allStarted, nil
	}
}

// tryStartRepeatIntervalFunc decide repeat rate of tryStartTaskFunc
// we repeat fast at beginning
// when all bindings are successfully started, we reduce the repeating rate
// S-shaped curve.
// Logistic Function 	https://en.wikipedia.org/wiki/Logistic_function
//
//	https://en.wikipedia.org/wiki/Generalised_logistic_function
func (b *SaramaKafkaBinder) tryStartRepeatIntervalFunc() loop.RepeatIntervalFunc {

	var fn func(int) time.Duration
	n := -1

	min := float64(b.properties.Binder.InitialHeartbeat)
	max := math.Max(min, float64(b.properties.Binder.WatchdogHeartbeat))
	mid := math.Max(1, b.properties.Binder.HeartbeatCurveMidpoint)
	k := math.Max(0.2, b.properties.Binder.HeartbeatCurveFactor)

	if float64(time.Minute) < max-min && mid >= 5 {
		fn = b.logisticModel(min, max, k, mid, time.Second)
	} else {
		fn = b.linearModel(min, max, mid)
	}

	return func(result interface{}, err error) time.Duration {
		switch allStarted := result.(type) {
		case bool:
			if allStarted {
				return time.Duration(b.properties.Binder.WatchdogHeartbeat)
			} else {
				ret := fn(n)
				n = n + 1
				//logger.Debugf("retry bindings in %dms", ret.Milliseconds())
				return ret
			}
		default:
			return time.Duration(b.properties.Binder.WatchdogHeartbeat)
		}
	}
}

// logisticModel returns delay function f(n) using logistic model
// Logistic Function 	https://en.wikipedia.org/wiki/Logistic_function
//
//	https://en.wikipedia.org/wiki/Generalised_logistic_function
func (b *SaramaKafkaBinder) logisticModel(min, max, k, n0 float64, y0 time.Duration) func(n int) time.Duration {
	// minK is calculated to make sure f(0) < min + y0 (first value is within y0 seconds of min value)
	minK := math.Log((max-min)/float64(y0)-1) / n0
	k = math.Max(k, minK)
	return func(n int) time.Duration {
		if n < 0 {
			return time.Duration(min)
		}
		return time.Duration((max-min)/(1+math.Exp(-k*(float64(n)-n0))) + min)
	}
}

// logisticModel returns delay function f(n) using linear model
func (b *SaramaKafkaBinder) linearModel(min, max, n0 float64) func(n int) time.Duration {
	return func(n int) time.Duration {
		return time.Duration((max-min)/n0/2*float64(n) + min)
	}
}
