package kafkatest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/kafka"
)

type MockedProducer struct {
	T string
	Recorder messageRecorder
}

func NewMockedProducer(topic string, recorder messageRecorder) *MockedProducer{
	return &MockedProducer{
		T:     topic,
		Recorder: recorder,
	}
}

func (p *MockedProducer) Topic() string {
	return p.T
}

func (p *MockedProducer) Send(_ context.Context, message interface{}, _ ...kafka.MessageOptions) error {
	p.Recorder.Record(&MessageRecord{
		Topic: p.T,
		Payload: message,
	})
	return nil
}

func(p *MockedProducer) ReadyCh() <-chan struct{} {
	// always ready
	ch := make(chan struct{}, 1)
	close(ch)
	return ch
}