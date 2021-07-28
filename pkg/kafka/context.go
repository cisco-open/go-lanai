package kafka

import "context"

type ProducerFactory interface {
	NewProducerWithTopic(topic string, options...ProducerOptions) (Producer, error)
}

type Producer interface {
	SendMessage(ctx context.Context, message interface{}, options...MessageOptions) error
}