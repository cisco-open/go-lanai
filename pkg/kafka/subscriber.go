package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/Shopify/sarama"
	"reflect"
)

type saramaSubscriber struct {
	topic      string
	brokers    []string
	config     *consumerConfig
	dispatcher *saramaDispatcher
	consumer   sarama.Consumer
	partitions []sarama.PartitionConsumer
	cancelFunc context.CancelFunc
}

func newSaramaSubscriber(topic string, addrs []string, config *consumerConfig) (*saramaSubscriber, error) {
	return &saramaSubscriber{
		topic:   topic,
		brokers: addrs,
		config:  config,
		dispatcher: newSaramaDispatcher(),
	}, nil
}

func (s *saramaSubscriber) Start(ctx context.Context) error {
	consumer, e := sarama.NewConsumer(s.brokers, s.config.Config)
	if e != nil {
		return translateSaramaBindingError(e, e.Error())
	}

	parts, e := consumer.Partitions(s.topic)
	if e != nil {
		return translateSaramaBindingError(e, e.Error())
	}

	s.partitions = make([]sarama.PartitionConsumer, len(parts))
	for i, p := range parts {
		s.partitions[i], e = consumer.ConsumePartition(s.topic, p, sarama.OffsetNewest)
		if e != nil {
			return translateSaramaBindingError(e, e.Error())
		}
	}

	cancelCtx, cancelFunc := context.WithCancel(ctx)
	go s.handlePartitions(cancelCtx, s.partitions)
	s.cancelFunc = cancelFunc
	return nil
}

func (s *saramaSubscriber) Close() error {
	defer func() {
		if s.cancelFunc != nil {
			s.cancelFunc()
			s.cancelFunc = nil
		}
	}()
	if e := s.consumer.Close(); e != nil {
		return NewKafkaError(ErrorCodeIllegalState, "error when closing subscriber: %v", e)
	}

	// cleanup
	s.consumer = nil
	s.partitions = nil
	return nil
}

func (s *saramaSubscriber) AddHandler(handlerFunc MessageHandlerFunc, opts...DispatchOptions) error {
	return s.dispatcher.addHandler(handlerFunc, opts)
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
		if !ok {
			logger.WithContext(ctx).Warnf("unrecognized object received from subscriber of partition [%d]: %T", chosen-1, val.Interface())
			continue
		}
		childCtx := utils.MakeMutableContext(ctx)
		s.logMessage(childCtx, msg)
		go s.handleMessage(childCtx, msg)
	}
}

// handleMessage intended to run in separate goroutine
func (s *saramaSubscriber) handleMessage(ctx context.Context, raw *sarama.ConsumerMessage) {
	if e := s.dispatcher.dispatch(ctx, raw); e != nil {
		logger.WithContext(ctx).Warnf("failed to handle message: %v", e)
	}
}

func (s *saramaSubscriber) logMessage(ctx context.Context, msg *sarama.ConsumerMessage) {
	logger.WithContext(ctx).Debugf("[RECV] [%s] Partition[%d] Offset[%d]: Length=%dB Key=%x", msg.Topic, msg.Partition, msg.Offset, len(msg.Value), msg.Key)
}