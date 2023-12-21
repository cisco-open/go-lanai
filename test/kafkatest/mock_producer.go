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
