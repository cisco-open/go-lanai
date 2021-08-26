package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"errors"
	"github.com/Shopify/sarama"
)

type saramaProducer struct {
	topic        string
	keyEncoder   Encoder
	msgLogger    MessageLogger
	interceptors []ProducerInterceptor
	syncProducer sarama.SyncProducer
}

func newSaramaProducer(topic string, addrs []string, config *producerConfig) (*saramaProducer, error) {
	c := *config //make a copy so that we don't change the original config
	//sync producer must have these two properties set to true
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true
	c.Producer.Partitioner = func(topic string) sarama.Partitioner {
		return sarama.NewRandomPartitioner(topic)
	}

	internal, err := sarama.NewSyncProducer(addrs, c.Config)
	if err != nil {
		return nil, translateSaramaBindingError(err, "unable to create producer: %v", err)
	}

	order.SortStable(config.interceptors, order.OrderedFirstCompare)
	p := &saramaProducer{
		topic:        topic,
		keyEncoder:   config.keyEncoder,
		msgLogger:    config.msgLogger,
		interceptors: config.interceptors,
		syncProducer: internal,
	}
	return p, nil
}

func (p *saramaProducer) Send(ctx context.Context, message interface{}, options ...MessageOptions) (err error) {
	msgCtx := p.prepare(ctx, message)
	if msgCtx.Payload == nil {
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
		Headers:  p.convertHeaders(msgCtx.Headers),
		Value:    msgCtx.Payload.(sarama.Encoder),
		Key:      newSaramaEncoder(msgCtx.Key, p.keyEncoder),
		Metadata: msgCtx,
	}

	// do send
	switch msgCtx.Mode {
	case modeSync:
		partition, offset, e := p.syncProducer.SendMessage(saramaMessage)
		// apply finalizers
		msgCtx, err = p.finalizeMessage(msgCtx, partition, offset, e)
	default:
		err = ErrorSubTypeIllegalProducerUsage.WithMessage("%v Mode is not supported", msgCtx.Mode)
	}
	return
}

func (p *saramaProducer) Close() error {
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

func (p *saramaProducer) finalizeMessage(msgCtx *MessageContext, partition int32, offset int64, err error) (*MessageContext, error) {
	for _, interceptor := range p.interceptors {
		switch finalizer := interceptor.(type) {
		case ProducerMessageFinalizer:
			msgCtx, err = finalizer.Finalize(msgCtx, partition, offset, err)
		}
	}
	if err == nil {
		return msgCtx, nil
	}

	switch {
	case errors.Is(err, ErrorCategoryKafka):
		return msgCtx, err
	default:
		return msgCtx, NewKafkaError(ErrorSubTypeCodeProducerGeneral, err.Error(), err)
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
