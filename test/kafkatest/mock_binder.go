// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package kafkatest

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/kafka"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"go.uber.org/fx"
	"sync"
)

type mockedBinderOut struct {
	fx.Out
	Binder   kafka.Binder
	Mock     *MockedBinder
	Recorder MessageRecorder
	Mocker   MessageMocker
}

func provideMockedBinder() mockedBinderOut {
	mock := MockedBinder{
		producers:   make(map[string]*MockedProducer),
		subscribers: make(map[string]*MockedSubscriber),
		consumers:   make(map[string]map[string]*MockedConsumer),
	}
	return mockedBinderOut{
		Binder:   &mock,
		Mock:     &mock,
		Recorder: &mock,
		Mocker: &mock,
	}
}

// MockedBinder implements kafka.Binder and messageRecorder
type MockedBinder struct {
	mtx         sync.Mutex
	producers   map[string]*MockedProducer
	subscribers map[string]*MockedSubscriber
	consumers   map[string]map[string]*MockedConsumer
	recordings  []*MessageRecord
}

func (b *MockedBinder) Produce(topic string, _ ...kafka.ProducerOptions) (kafka.Producer, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	ret, ok := b.producers[topic]
	if !ok {
		ret = NewMockedProducer(topic, b)
		b.producers[topic] = ret
	}
	return ret, nil
}

func (b *MockedBinder) Subscribe(topic string, _ ...kafka.ConsumerOptions) (kafka.Subscriber, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	ret, ok := b.subscribers[topic]
	if !ok {
		ret = NewMockedSubscriber(topic)
		b.subscribers[topic] = ret
	}
	return ret, nil
}

func (b *MockedBinder) Consume(topic string, group string, _ ...kafka.ConsumerOptions) (kafka.GroupConsumer, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	grouped, ok := b.consumers[topic]
	if !ok {
		grouped = make(map[string]*MockedConsumer)
		b.consumers[topic] = grouped
	}
	ret, ok := grouped[group]
	if !ok {
		ret = NewMockedConsumer(topic, group)
		grouped[group] = ret
	}
	return ret, nil
}

func (b *MockedBinder) ListTopics() []string {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	topics := utils.NewStringSet()
	for k := range b.producers {
		topics.Add(k)
	}
	for k := range b.subscribers {
		topics.Add(k)
	}
	for k := range b.consumers {
		topics.Add(k)
	}
	return topics.Values()
}

func (b *MockedBinder) Reset() {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.recordings = nil
}

func (b *MockedBinder) Records(topic string) (ret []*MessageRecord) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	ret = make([]*MessageRecord, 0, len(b.recordings))
	for _, r := range b.recordings {
		if r.Topic == topic {
			ret = append(ret, r)
		}
	}
	return
}

func (b *MockedBinder) AllRecords() (ret []*MessageRecord) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	ret = make([]*MessageRecord, len(b.recordings))
	copy(ret, b.recordings)
	return
}

func (b *MockedBinder) Record(record *MessageRecord) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.recordings = append(b.recordings, record)
}

func (b *MockedBinder) Mock(ctx context.Context, topic string, msg *kafka.Message) error {
	msgCtx := b.mockMessageContext(ctx, topic, msg)
	b.mtx.Lock()
	defer b.mtx.Unlock()
	dispatcher, ok := b.subscribers[topic]
	if !ok {
		return nil
	}
	return dispatcher.Dispatch(msgCtx)
}

func (b *MockedBinder) MockWithGroup(ctx context.Context, topic, group string, msg *kafka.Message) error {
	msgCtx := b.mockMessageContext(ctx, topic, msg)
	b.mtx.Lock()
	defer b.mtx.Unlock()
	consumers, ok := b.consumers[topic]
	if !ok {
		return nil
	}
	dispatcher, ok := consumers[group]
	if !ok {
		return nil
	}
	return dispatcher.Dispatch(msgCtx)
}

func (b *MockedBinder) mockMessageContext(ctx context.Context, topic string, msg *kafka.Message) *kafka.MessageContext {
	return &kafka.MessageContext{
		Context:    ctx,
		Source:     b,
		Topic:      topic,
		Message:    *msg,
		RawMessage: msg,
	}
}
