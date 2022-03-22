package kafkatest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/kafka"
)

type MockedConsumer struct {
	T  string
	G string
}

func NewMockedConsumer(topic, group string) *MockedConsumer {
	return &MockedConsumer{
		T: topic,
		G: group,
	}
}

func (c *MockedConsumer) Topic() string {
	return c.T
}

func (c *MockedConsumer) Group() string {
	return c.G
}

func (c *MockedConsumer) AddHandler(handlerFunc kafka.MessageHandlerFunc, opts ...kafka.DispatchOptions) error {
	// noop for now
	return nil
}
