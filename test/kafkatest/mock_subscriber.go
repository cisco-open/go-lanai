package kafkatest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/kafka"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

type MockedSubscriber struct {
	T     string
	Parts []int32
}

func NewMockedSubscriber(topic string) *MockedSubscriber {
	return &MockedSubscriber{
		T:     topic,
		Parts: []int32{int32(utils.RandomIntN(0xffff))},
	}
}

func (s *MockedSubscriber) Topic() string {
	return s.T
}

func (s *MockedSubscriber) Partitions() []int32 {
	return []int32{}
}

func (s *MockedSubscriber) AddHandler(handlerFunc kafka.MessageHandlerFunc, opts ...kafka.DispatchOptions) error {
	// noop
	return nil
}


