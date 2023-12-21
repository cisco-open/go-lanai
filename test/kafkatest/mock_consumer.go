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
