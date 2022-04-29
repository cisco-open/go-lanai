package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"sync"
)

type saramaProducer struct {
	sync.RWMutex
	topic        string
	brokers      []string
	config       *bindingConfig
	keyEncoder   Encoder
	msgLogger    MessageLogger
	interceptors []ProducerMessageInterceptor
	syncProducer sarama.SyncProducer
	closed       bool
}

func newSaramaProducer(topic string, addrs []string, config *bindingConfig) (*saramaProducer, error) {
	//sync producer must have these two properties set to true
	config.sarama.Producer.Return.Successes = true
	config.sarama.Producer.Return.Errors = true
	config.sarama.Producer.Partitioner = func(topic string) sarama.Partitioner {
		return sarama.NewRandomPartitioner(topic)
	}

	order.SortStable(config.producer.interceptors, order.OrderedFirstCompare)
	p := &saramaProducer{
		topic:        topic,
		brokers:      addrs,
		config:       config,
		keyEncoder:   config.producer.keyEncoder,
		msgLogger:    config.msgLogger,
		interceptors: config.producer.interceptors,
	}
	return p, nil
}

func (p *saramaProducer) Topic() string {
	return p.topic
}

func (p *saramaProducer) Send(ctx context.Context, message interface{}, options ...MessageOptions) (err error) {
	var syncProducer sarama.SyncProducer
	p.RLock()
	syncProducer = p.syncProducer
	p.RUnlock()

	if syncProducer == nil {
		return NewKafkaError(ErrorSubTypeCodeIllegalProducerUsage, fmt.Sprintf(`producer for topic "%s" is not started yet`, p.topic))
	}

	msgCtx := p.prepare(ctx, message)
	if msgCtx.Message.Payload == nil {
		return nil
	}

	// apply options
	for _, optionFunc := range options {
		optionFunc(&msgCtx.messageConfig)
	}

	// apply interceptors
	for _, interceptor := range p.interceptors {
		if msgCtx, err = interceptor.Intercept(msgCtx); err != nil {
			return ErrorSubTypeProducerGeneral.WithMessage("producer interceptor error: %v", err)
		}
	}

	// initialize sarama message
	saramaMessage := &sarama.ProducerMessage{
		Topic:    p.topic,
		Headers:  p.convertHeaders(msgCtx.Message.Headers),
		Value:    msgCtx.Message.Payload.(sarama.Encoder),
		Key:      newSaramaEncoder(msgCtx.Key, p.keyEncoder),
		Metadata: msgCtx,
	}
	msgCtx.RawMessage = saramaMessage

	// do send
	switch msgCtx.Mode {
	case modeSync:
		partition, offset, e := syncProducer.SendMessage(saramaMessage)
		// apply finalizers
		err = p.finalizeSend(msgCtx, partition, offset, e)
	default:
		err = ErrorSubTypeIllegalProducerUsage.WithMessage("%v Mode is not supported", msgCtx.Mode)
	}
	return
}

func (p *saramaProducer) Start(_ context.Context) error {
	p.Lock()
	defer p.Unlock()
	switch {
	case p.closed:
		return ErrorStartClosedBinding.WithMessage("cannot re-start a closed producer [%s]", p.topic)
	case p.syncProducer != nil:
		return nil
	}
	internal, e := sarama.NewSyncProducer(p.brokers, &p.config.sarama)
	if e != nil {
		return translateSaramaBindingError(e, "unable to start producer: %v", e)
	}
	p.syncProducer = internal
	return nil
}

func (p *saramaProducer) Close() error {
	p.Lock()
	defer p.Unlock()
	if p.syncProducer == nil {
		return nil
	}
	if e := p.syncProducer.Close(); e != nil {
		return NewKafkaError(ErrorCodeIllegalState, "error when closing producer: %v", e)
	}
	p.closed = true
	return nil
}

func (p *saramaProducer) Closed() bool {
	p.Lock()
	defer p.Unlock()
	return p.closed
}

func (p *saramaProducer) prepare(ctx context.Context, v interface{}) *MessageContext {
	msgCtx := MessageContext{
		Context:       ctx,
		Topic:         p.topic,
		messageConfig: defaultMessageConfig(),
		Source:        p,
	}
	switch m := v.(type) {
	case *Message:
		msgCtx.Message = *m
	case Message:
		msgCtx.Message = m
	default:
		msgCtx.Message = Message{
			Headers: Headers{},
			Payload: v,
		}
	}
	if msgCtx.Message.Headers == nil {
		msgCtx.Message.Headers = Headers{}
	}
	return &msgCtx
}

func (p *saramaProducer) finalizeSend(msgCtx *MessageContext, partition int32, offset int64, err error) error {

	p.msgLogger.LogSentMessage(msgCtx.Context, msgCtx.RawMessage)

	for _, interceptor := range p.interceptors {
		switch finalizer := interceptor.(type) {
		case ProducerMessageFinalizer:
			msgCtx, err = finalizer.Finalize(msgCtx, partition, offset, err)
		}
	}
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ErrorCategoryKafka):
		return err
	default:
		return NewKafkaError(ErrorSubTypeCodeProducerGeneral, err.Error(), err)
	}
}

func (p *saramaProducer) convertHeaders(headers Headers) (ret []sarama.RecordHeader) {
	if headers == nil {
		return
	}
	ret = make([]sarama.RecordHeader, len(headers))
	var i int
	for k, v := range headers {
		ret[i] = sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		}
		i++
	}
	return
}
