package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"sync"
)

type saramaGroupConsumer struct {
	topic      string
	group      string
	brokers    []string
	config     *consumerConfig
	dispatcher *saramaDispatcher
	msgLogger  MessageLogger
	startOnce  sync.Once
	consumer   sarama.ConsumerGroup
	cancelFunc context.CancelFunc
}

func newSaramaGroupConsumer(topic string, group string, addrs []string, config *consumerConfig) (*saramaGroupConsumer, error) {
	if group == "" {
		return nil, ErrorSubTypeBindingInternal.WithMessage("group is required and cannot be empty")
	}

	config.Config.Consumer.Return.Errors = true
	return &saramaGroupConsumer{
		topic:      topic,
		group:      group,
		brokers:    addrs,
		config:     config,
		dispatcher: newSaramaDispatcher(),
		msgLogger:  config.msgLogger,
	}, nil
}

func (g *saramaGroupConsumer) Topic() string {
	return g.topic
}

func (g *saramaGroupConsumer) Group() string {
	return g.group
}

func (g *saramaGroupConsumer) Start(ctx context.Context) (err error) {
	g.startOnce.Do(func() {
		group, e := sarama.NewConsumerGroup(g.brokers, g.group, g.config.Config)
		if e != nil {
			err = translateSaramaBindingError(e, e.Error())
			return
		}
		g.consumer = group

		cancelCtx, cancelFunc := context.WithCancel(ctx)
		go g.monitorGroupErrors(cancelCtx, g.consumer)
		go g.handleGroup(cancelCtx, g.consumer)
		g.cancelFunc = cancelFunc
	})
	return
}

func (g *saramaGroupConsumer) Close() error {
	defer func() {
		if g.cancelFunc != nil {
			g.cancelFunc()
			g.cancelFunc = nil
		}
	}()
	if e := g.consumer.Close(); e != nil {
		return NewKafkaError(ErrorCodeIllegalState, "error when closing subscriber: %v", e)
	}

	// cleanup
	g.consumer = nil
	return nil
}

func (g *saramaGroupConsumer) AddHandler(handlerFunc MessageHandlerFunc, opts ...DispatchOptions) error {
	return g.dispatcher.addHandler(handlerFunc, opts)
}

// monitorGroupErrors should be run in separate goroutine
func (g *saramaGroupConsumer) monitorGroupErrors(ctx context.Context, group sarama.ConsumerGroup) {
	for {
		select {
		case err := <-group.Errors():
			logger.WithContext(ctx).Warnf("Consumer Group Error: %v", err)
		case <-ctx.Done():
			return
		}
	}
}

// handleGroup should be run in separate goroutine
func (g *saramaGroupConsumer) handleGroup(ctx context.Context, group sarama.ConsumerGroup) {
	for {
		gh := saramaGroupHandler{
			dispatcher: g.dispatcher,
			msgLogger:  g.msgLogger,
		}

		// `Consume` should be called inside an infinite loop, when a server-side re-balance happens, the consumer session will need to be recreated to get the new claims
		if e := group.Consume(ctx, []string{g.topic}, gh); e != nil {
			logger.WithContext(ctx).Warnf("Consumer Group Error: %v", e)
		}
	}
}

type saramaGroupHandler struct {
	dispatcher *saramaDispatcher
	msgLogger  MessageLogger
}

func (s saramaGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	// TODO review this
	for group, parts := range session.Claims() {
		logger.WithContext(session.Context()).
			Debugf("Consumer [%s] joined group [%s] with partitions %v", session.MemberID(), group, parts)
	}
	return nil
}

func (s saramaGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	// TODO review this
	for group, parts := range session.Claims() {
		logger.WithContext(session.Context()).
			Debugf("Consumer [%s] left group [%s] releasing partitions %v", session.MemberID(), group, parts)
	}
	return nil
}

// ConsumeClaim is run in separate goroutine
func (s saramaGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			s.msgLogger.LogReceivedMessage(session.Context(), msg)
			go s.handleMessage(session.Context(), session, msg)
		case <-session.Context().Done():
			return nil
		}
	}
}

// handleMessage intended to run in separate goroutine
func (s saramaGroupHandler) handleMessage(ctx context.Context, session sarama.ConsumerGroupSession, raw *sarama.ConsumerMessage) {
	if e := s.dispatcher.dispatch(ctx, raw); e != nil {
		logger.WithContext(ctx).Warnf("failed to handle message: %v", e)
		session.ResetOffset(raw.Topic, raw.Partition, raw.Offset, e.Error())
		return
	}
	session.MarkMessage(raw, "")
}
