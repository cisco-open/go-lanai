package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/Shopify/sarama"
	"reflect"
	"sync"
)

type saramaSubscriber struct {
	sync.Mutex
	topic      string
	brokers    []string
	config     *consumerConfig
	dispatcher *saramaDispatcher
	provisioner *saramaTopicProvisioner
	started    bool
	consumer   sarama.Consumer
	partitions []int32
	cancelFunc context.CancelFunc
}

func newSaramaSubscriber(topic string, addrs []string, config *consumerConfig, provisioner *saramaTopicProvisioner) (*saramaSubscriber, error) {
	return &saramaSubscriber{
		topic:      topic,
		brokers:    addrs,
		config:     config,
		dispatcher: newSaramaDispatcher(config),
		provisioner: provisioner,
	}, nil
}

func (s *saramaSubscriber) Topic() string {
	return s.topic
}

func (s *saramaSubscriber) Partitions() []int32 {
	return s.partitions
}

func (s *saramaSubscriber) Start(ctx context.Context) (err error) {
	s.Lock()
	defer s.Unlock()
	defer func() {
		if err == nil {
			s.started = true
		}
	}()

	if s.started {
		return nil
	}

	if ok, e := s.provisioner.topicExists(s.topic); e != nil || !ok {
		return NewKafkaError(ErrorCodeIllegalState, fmt.Sprintf(`topic "%s" does not exists`, s.topic))
	}

	var e error
	if s.consumer, e = sarama.NewConsumer(s.brokers, s.config.Config); e != nil {
		err = translateSaramaBindingError(e, e.Error())
		return
	}

	if s.partitions, e = s.consumer.Partitions(s.topic); e != nil {
		err = translateSaramaBindingError(e, e.Error())
		return
	}

	partitionConsumers := make([]sarama.PartitionConsumer, len(s.partitions))
	for i, p := range s.partitions {
		if partitionConsumers[i], e = s.consumer.ConsumePartition(s.topic, p, sarama.OffsetNewest); e != nil {
			err = translateSaramaBindingError(e, e.Error())
			return
		}
	}

	cancelCtx, cancelFunc := context.WithCancel(ctx)
	go s.handlePartitions(cancelCtx, partitionConsumers)
	s.cancelFunc = cancelFunc
	return
}

func (s *saramaSubscriber) Close() error {
	s.Lock()
	defer s.Unlock()
	defer func() {s.started = false}()

	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}

	if s.consumer == nil {
		return nil
	}

	if e := s.consumer.Close(); e != nil {
		return NewKafkaError(ErrorCodeIllegalState, "error when closing subscriber: %v", e)
	}

	return nil
}

func (s *saramaSubscriber) AddHandler(handlerFunc MessageHandlerFunc, opts ...DispatchOptions) error {
	return s.dispatcher.addHandler(handlerFunc, s.config, opts)
}

// handlePartitions intended to run in separate goroutine
func (s *saramaSubscriber) handlePartitions(ctx context.Context, partitions []sarama.PartitionConsumer) {
	cases := make([]reflect.SelectCase, len(partitions)+1)
	cases[0] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}
	for i, pc := range partitions {
		cases[i+1] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(pc.Messages()),
		}
	}

	for {
		chosen, val, ok := reflect.Select(cases)
		if !ok || chosen == 0 {
			// channel closed or Done channel received
			break
		}
		msg, ok := val.Interface().(*sarama.ConsumerMessage)
		if !ok || msg == nil {
			logger.WithContext(ctx).Warnf("unrecognized object received from subscriber of partition [%d]: %T", chosen-1, val.Interface())
			continue
		}
		childCtx := utils.MakeMutableContext(ctx)
		go s.handleMessage(childCtx, msg)
	}
}

// handleMessage intended to run in separate goroutine
func (s *saramaSubscriber) handleMessage(ctx context.Context, raw *sarama.ConsumerMessage) {
	if e := s.dispatcher.dispatch(ctx, raw, s); e != nil {
		logger.WithContext(ctx).Warnf("failed to handle message: %v", e)
	}
}
