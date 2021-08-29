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
	sync.Mutex
	topic        string
	brokers      []string
	config       *producerConfig
	keyEncoder   Encoder
	msgLogger    MessageLogger
	interceptors []ProducerMessageInterceptor
	syncProducer sarama.SyncProducer
}

func newSaramaProducer(topic string, addrs []string, config *producerConfig) (*saramaProducer, error) {
	//sync producer must have these two properties set to true
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Partitioner = func(topic string) sarama.Partitioner {
		return sarama.NewRandomPartitioner(topic)
	}

	order.SortStable(config.interceptors, order.OrderedFirstCompare)
	p := &saramaProducer{
		topic:        topic,
		brokers:      addrs,
		config:       config,
		keyEncoder:   config.keyEncoder,
		msgLogger:    config.msgLogger,
		interceptors: config.interceptors,
	}
	return p, nil
}

func (p *saramaProducer) Topic() string {
	return p.topic
}

func (p *saramaProducer) Send(ctx context.Context, message interface{}, options ...MessageOptions) (err error) {
	// note: if necessary, we would consider use RWLock
	if p.syncProducer == nil {
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
		partition, offset, e := p.syncProducer.SendMessage(saramaMessage)
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
	internal, e := sarama.NewSyncProducer(p.brokers, &p.config.Config)
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
	return nil
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
