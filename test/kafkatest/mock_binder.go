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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/kafka"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"go.uber.org/fx"
	"sync"
)

type mockedBinderOut struct {
	fx.Out
	Binder   kafka.Binder
	Mock     *MockedBinder
	Recorder MessageRecorder
}

func provideMockedBinder() mockedBinderOut {
	mock := MockedBinder{
		producers:   make(map[string]kafka.Producer),
		subscribers: make(map[string]kafka.Subscriber),
		consumers:   make(map[string]kafka.GroupConsumer),
	}
	return mockedBinderOut{
		Binder:   &mock,
		Mock:     &mock,
		Recorder: &mock,
	}
}

// MockedBinder implements kafka.Binder and messageRecorder
type MockedBinder struct {
	mtx         sync.Mutex
	producers   map[string]kafka.Producer
	subscribers map[string]kafka.Subscriber
	consumers   map[string]kafka.GroupConsumer
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
	ret, ok := b.subscribers[topic]
	if !ok {
		ret = NewMockedSubscriber(topic)
		b.subscribers[topic] = ret
	}
	return ret, nil
}

func (b *MockedBinder) Consume(topic string, group string, _ ...kafka.ConsumerOptions) (kafka.GroupConsumer, error) {
	ret, ok := b.consumers[topic]
	if !ok {
		ret = NewMockedConsumer(topic, group)
		b.consumers[topic] = ret
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
