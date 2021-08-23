package kafka

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
)

type SaramaProducer struct {
	topic        string
	keyEncoder   Encoder
	interceptors []ProducerInterceptor
	syncProducer sarama.SyncProducer
}

func newSaramaProducer(topic string, addrs []string, config *producerConfig) (*SaramaProducer, error) {
	c := *config //make a copy so that we don't change the original config
	//sync producer must have these two properties set to true
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true

	internal, err := sarama.NewSyncProducer(addrs, c.Config)
	if err != nil {
		return nil, err
	}

	order.SortStable(config.interceptors, order.OrderedFirstCompare)
	p := &SaramaProducer{
		topic:        topic,
		keyEncoder:   config.keyEncoder,
		interceptors: config.interceptors,
		syncProducer: internal,
	}
	return p, nil
}

func (p *SaramaProducer) SendMessage(ctx context.Context, message interface{}, options ...MessageOptions) (err error) {
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
			return
		}
	}

	// initialize sarama message
	saramaMessage := &sarama.ProducerMessage{
		Topic:   p.topic,
		Headers: p.convertHeaders(msgCtx.Headers),
		Value:   msgCtx.Payload.(sarama.Encoder),
		Key:     newSaramaEncoder(msgCtx.Key, p.keyEncoder),
		Metadata: msgCtx,
	}

	// do send
	switch msgCtx.Mode {
	case modeSync:
		partition, offset, e := p.syncProducer.SendMessage(saramaMessage)
		// apply finalizers
		msgCtx, err = p.finalizeMessage(msgCtx, partition, offset, e)
	default:
		err = errors.New(fmt.Sprintf("%v Mode is not supported", msgCtx.Mode))
	}
	return
}

func (p *SaramaProducer) Close() error {
	return p.syncProducer.Close()
}

func (p *SaramaProducer) prepare(ctx context.Context, v interface{}) *MessageContext {
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

func (p *SaramaProducer) finalizeMessage(msgCtx *MessageContext, partition int32, offset int64, err error) (*MessageContext, error) {
	for _, interceptor := range p.interceptors {
		switch finalizer := interceptor.(type) {
		case ProducerMessageFinalizer:
			msgCtx, err = finalizer.Finalize(msgCtx, partition, offset, err)
		}
	}
	return msgCtx, err
}

func (p *SaramaProducer) convertHeaders(headers Headers) (ret []sarama.RecordHeader) {
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
