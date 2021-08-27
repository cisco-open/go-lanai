package kafka

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"sync"
)

type saramaGroupConsumer struct {
	sync.Mutex
	topic       string
	group       string
	brokers     []string
	config      *consumerConfig
	dispatcher  *saramaDispatcher
	provisioner *saramaTopicProvisioner
	started     bool
	consumer    sarama.ConsumerGroup
	cancelFunc  context.CancelFunc
}

func newSaramaGroupConsumer(topic string, group string, addrs []string, config *consumerConfig, provisioner *saramaTopicProvisioner) (*saramaGroupConsumer, error) {
	if group == "" {
		return nil, ErrorSubTypeBindingInternal.WithMessage("group is required and cannot be empty")
	}

	//config.Consumer.Return.Errors = true
	return &saramaGroupConsumer{
		topic:       topic,
		group:       group,
		brokers:     addrs,
		config:      config,
		dispatcher:  newSaramaDispatcher(config),
		provisioner: provisioner,
	}, nil
}

func (g *saramaGroupConsumer) Topic() string {
	return g.topic
}

func (g *saramaGroupConsumer) Group() string {
	return g.group
}

func (g *saramaGroupConsumer) Start(ctx context.Context) (err error) {
	g.Lock()
	defer g.Unlock()
	defer func() {
		if err == nil {
			g.started = true
		}
	}()
	if g.started {
		return nil
	}

	if ok, e := g.provisioner.topicExists(g.topic); e != nil || !ok {
		return NewKafkaError(ErrorCodeIllegalState, fmt.Sprintf(`topic "%s" does not exists`, g.topic))
	}

	var e error
	g.consumer, e = sarama.NewConsumerGroup(g.brokers, g.group, g.config.Config)
	if e != nil {
		err = translateSaramaBindingError(e, e.Error())
		return
	}

	cancelCtx, cancelFunc := context.WithCancel(ctx)
	if g.config.Consumer.Return.Errors {
		go g.monitorGroupErrors(cancelCtx, g.consumer)
	}
	go g.handleGroup(cancelCtx, g.consumer)
	g.cancelFunc = cancelFunc
	return
}

func (g *saramaGroupConsumer) Close() error {
	g.Lock()
	defer g.Unlock()
	defer func() { g.started = false }()

	if g.cancelFunc != nil {
		g.cancelFunc()
		g.cancelFunc = nil
	}

	if g.consumer == nil {
		return nil
	}

	if e := g.consumer.Close(); e != nil {
		return NewKafkaError(ErrorCodeIllegalState, "error when closing group consumer: %v", e)
	}

	return nil
}

func (g *saramaGroupConsumer) AddHandler(handlerFunc MessageHandlerFunc, opts ...DispatchOptions) error {
	return g.dispatcher.addHandler(handlerFunc, g.config, opts)
}

// monitorGroupErrors should be run in separate goroutine
func (g *saramaGroupConsumer) monitorGroupErrors(ctx context.Context, group sarama.ConsumerGroup) {
	for {
		select {
		case e, ok := <-group.Errors():
			if !ok {
				return
			}
			if e == sarama.ErrClosedConsumerGroup {
				return
			}
			logger.WithContext(ctx).Warnf("Consumer Group Error: %v", e)
		case <-ctx.Done():
			return
		}
	}
}

// handleGroup should be run in separate goroutine
func (g *saramaGroupConsumer) handleGroup(ctx context.Context, group sarama.ConsumerGroup) {
	gh := saramaGroupHandler{
		owner:      g,
		dispatcher: g.dispatcher,
	}

	for {
		// `Consume` should be called inside an infinite loop, when a server-side re-balance happens, the consumer session will need to be recreated to get the new claims
		if e := group.Consume(ctx, []string{g.topic}, gh); e != nil {
			if e == sarama.ErrClosedConsumerGroup {
				return
			}
			logger.WithContext(ctx).Warnf("Consumer Error: %v", e)
		}
	}
}

type saramaGroupHandler struct {
	owner      *saramaGroupConsumer
	dispatcher *saramaDispatcher
}

func (h saramaGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	for topic, parts := range session.Claims() {
		logger.WithContext(session.Context()).
			Debugf("Consumer joined group [%s] Topic=[%s] Partitions=%v MemberID=[%s]", h.owner.group, topic, parts, session.MemberID())
	}
	return nil
}

func (h saramaGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	for topic, parts := range session.Claims() {
		logger.WithContext(session.Context()).
			Debugf("Consumer left group [%s] Topic=[%s] Partitions=%v MemberID=[%s]", h.owner.group, topic, parts, session.MemberID())
	}
	return nil
}

// ConsumeClaim is run in separate goroutine
func (h saramaGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			go h.handleMessage(session.Context(), session, msg)
		case <-session.Context().Done():
			return nil
		}
	}
}

// handleMessage intended to run in separate goroutine
func (h saramaGroupHandler) handleMessage(ctx context.Context, session sarama.ConsumerGroupSession, raw *sarama.ConsumerMessage) {
	if e := h.dispatcher.dispatch(ctx, raw, h.owner); e != nil {
		logger.WithContext(ctx).Warnf("failed to handle message: %v", e)
		session.ResetOffset(raw.Topic, raw.Partition, raw.Offset, e.Error())
		return
	}
	session.MarkMessage(raw, "")
}
