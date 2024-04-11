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
	"github.com/cisco-open/go-lanai/pkg/kafka"
	"github.com/cisco-open/go-lanai/pkg/utils"
)

type MockedSubscriber struct {
	kafka.Dispatcher
	T          string
	Parts      []int32
}

func NewMockedSubscriber(topic string) *MockedSubscriber {
	return &MockedSubscriber{
		Dispatcher: kafka.Dispatcher{
			Logger: messageLogger,
		},
		T:     topic,
		Parts: []int32{int32(utils.RandomIntN(0xffff))},
	}
}

func (s *MockedSubscriber) Topic() string {
	return s.T
}

func (s *MockedSubscriber) Partitions() []int32 {
	return s.Parts
}


