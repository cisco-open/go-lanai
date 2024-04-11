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
	"github.com/cisco-open/go-lanai/pkg/log"
)

type MessageRecord struct {
	Topic   string
	Payload interface{}
}

// MessageRecorder interface for retrieve messages produced by MockedProducer
type MessageRecorder interface {
	Reset()
	Records(topic string) []*MessageRecord
	AllRecords() []*MessageRecord
}

type messageRecorder interface {
	MessageRecorder
	Record(msg *MessageRecord)
}

// MessageMocker interface for mocking incoming messages.
type MessageMocker interface {
	Mock(ctx context.Context, topic string, msg *kafka.Message) error
	MockWithGroup(ctx context.Context, topic, group string, msg *kafka.Message) error
}

type msgLogger struct {
	logger log.ContextualLogger
	level  log.LoggingLevel
}

func (l msgLogger) WithLevel(level log.LoggingLevel) kafka.MessageLogger {
	return msgLogger{
		logger: l.logger,
		level:  level,
	}
}

func (l msgLogger) LogSentMessage(ctx context.Context, msg interface{}) {
	l.logger.WithContext(ctx).WithLevel(l.level).Printf(`Sent: %v`, msg)
}

func (l msgLogger) LogReceivedMessage(ctx context.Context, msg interface{}) {
	l.logger.WithContext(ctx).WithLevel(l.level).Printf(`Received: %v`, msg)
}
