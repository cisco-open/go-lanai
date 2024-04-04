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

package examples

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/kafka"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/kafkatest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

const (
	ExampleTopic = "TEST_EVENTS"
	ExampleGroup = "TESTS"
)

/*************************
	Setup
 *************************/

type tsDI struct {
	fx.In
	Binder kafka.Binder
}

type TestMessage struct {
	Int    int
	String string
}

type TestService struct {
	producer kafka.Producer
	lastMsg  *TestMessage
}

func NewTestService(di tsDI) *TestService {
	svc := &TestService{}
	p, e := di.Binder.Produce(ExampleTopic, kafka.BindingName("test"), kafka.RequireLocalAck())
	if e != nil {
		panic(e)
	}
	svc.producer = p

	s, e := di.Binder.Subscribe(ExampleTopic)
	if e != nil {
		panic(e)
	}
	if e := s.AddHandler(svc.Handle); e != nil {
		panic(e)
	}

	c, e := di.Binder.Consume(ExampleTopic, ExampleGroup)
	if e != nil {
		panic(e)
	}
	if e := c.AddHandler(svc.Handle); e != nil {
		panic(e)
	}
	return svc
}

func (s *TestService) GenerateSomeMessages(ctx context.Context, count int) error {
	for i := 0; i < count; i++ {
		e := s.producer.Send(ctx, &TestMessage{
			Int:    i,
			String: fmt.Sprintf("Message-%d", i),
		}, kafka.WithKey(uuid.New()))
		if e != nil {
			return e
		}
	}
	return nil
}

func (s *TestService) Handle(_ context.Context, msg *TestMessage, _ *kafka.Message) error {
	s.lastMsg = msg
	return nil
}

/*************************
	Tests
 *************************/

type testDI struct {
	fx.In
	Service  *TestService
	Recorder kafkatest.MessageRecorder
	Mocker   kafkatest.MessageMocker
}

func TestMockedBinder(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		kafkatest.WithMockedBinder(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(NewTestService),
		),
		test.GomegaSubTest(SubTestExampleProducerRecording(di), "ExampleProducerRecording"),
		test.GomegaSubTest(SubTestExampleSubscriberMessageMocking(di), "ExampleSubscriberMessageMocking"),
		test.GomegaSubTest(SubTestExampleConsumerMessageMocking(di), "ExampleConsumerMessageMocking"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestExampleProducerRecording(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// don't forget to reset recorder
		di.Recorder.Reset()

		var e error
		// Do something that producing messages
		e = di.Service.GenerateSomeMessages(ctx, 3)
		g.Expect(e).To(Succeed(), "functions using producers shouldn't fail")

		// validate recorded messages
		actual := di.Recorder.Records(ExampleTopic)
		g.Expect(actual).To(HaveLen(3), "recorded messages should have correct length")
		for i, record := range actual {
			g.Expect(record.Payload).To(BeAssignableToTypeOf(&TestMessage{}), "recorded message at [%d] should have correct type", i)
			msg := record.Payload.(*TestMessage)
			g.Expect(msg.Int).To(BeEquivalentTo(i), "recorded message at [%d] should have correct Int field", i)
			g.Expect(msg.String).To(BeEquivalentTo(fmt.Sprintf("Message-%d", i)), "recorded message at [%d] should have correct String field", i)
		}
	}
}

func SubTestExampleSubscriberMessageMocking(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// reset service
		di.Service.lastMsg = nil
		// mock some message
		mock := &kafka.Message{
			Payload: &TestMessage{
				Int:    10,
				String: "mya",
			},
		}
		e := di.Mocker.Mock(ctx, ExampleTopic, mock)
		g.Expect(e).To(Succeed(), "mocking incoming message should not fail")
		// verify
		g.Expect(di.Service.lastMsg).ToNot(BeNil(), "handler should already handled the message")
	}
}

func SubTestExampleConsumerMessageMocking(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// reset service
		di.Service.lastMsg = nil
		// mock some message
		mock := &kafka.Message{
			Payload: &TestMessage{
				Int:    10,
				String: "mya",
			},
		}
		e := di.Mocker.MockWithGroup(ctx, ExampleTopic, ExampleGroup, mock)
		g.Expect(e).To(Succeed(), "mocking incoming message should not fail")
		// verify
		g.Expect(di.Service.lastMsg).ToNot(BeNil(), "handler should already handled the message")
	}
}
