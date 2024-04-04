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
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/kafka"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

const (
	TestTopic = "TEST_EVENTS"
	TestGroup = "TESTS"
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
	received []*TestMessage
}

func NewTestService(di tsDI) (*TestService, error) {
	svc := &TestService{}
	p, e := di.Binder.Produce(TestTopic, kafka.BindingName("test"), kafka.RequireLocalAck())
	if e != nil {
		return nil, e
	}
	svc.producer = p

	s, e := di.Binder.Subscribe(TestTopic)
	if e != nil {
		return nil, e
	}
	if e := s.AddHandler(svc.Handle); e != nil {
		return nil, e
	}

	c, e := di.Binder.Consume(TestTopic, TestGroup)
	if e != nil {
		return nil, e
	}
	if e := c.AddHandler(svc.Handle); e != nil {
		return nil, e
	}
	return svc, nil
}

func (s *TestService) GenerateSomeMessages(ctx context.Context, count int) error {
	timoutCtx, cancelFn := context.WithTimeout(ctx, 1*time.Second)
	defer cancelFn()
	select {
	case <-s.producer.ReadyCh():
	case <-timoutCtx.Done():
		return fmt.Errorf("producer is not ready")
	}
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
	s.received = append(s.received, msg)
	return nil
}

/*************************
	Tests
 *************************/

type testDI struct {
	fx.In
	Recorder MessageRecorder `optional:"true"`
	Mocker   MessageMocker   `optional:"true"`
	Service  *TestService    `optional:"true"`
	Binder   kafka.Binder
}

func TestMockedBinder(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithMockedBinder(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(NewTestService),
		),
		test.GomegaSubTest(SubTestProducerRecording(di), "ProducerRecording"),
		test.GomegaSubTest(SubTestSubscriber(di), "Subscriber"),
		test.GomegaSubTest(SubTestConsumer(di), "Consumer"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestProducerRecording(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).NotTo(BeNil(), "MessageRecorder should be available")
		g.Expect(di.Service).NotTo(BeNil(), "TestService should be available")
		g.Expect(di.Service.producer.Topic()).To(Equal(TestTopic), "TestService.producer's Topic() should be correct")

		var e error
		di.Recorder.Reset()
		e = di.Service.GenerateSomeMessages(ctx, 3)
		g.Expect(e).To(Succeed(), "functions using producers shouldn't fail")
		assertRecordedMessages(t, g, di.Recorder.Records(TestTopic), 3, true)

		// do it again without reset
		e = di.Service.GenerateSomeMessages(ctx, 3)
		g.Expect(e).To(Succeed(), "functions using producers shouldn't fail")
		assertRecordedMessages(t, g, di.Recorder.Records(TestTopic), 6, false)

		// all records
		assertRecordedMessages(t, g, di.Recorder.AllRecords(), 6, false)

		// validate reset
		di.Recorder.Reset()
		assertRecordedMessages(t, g, di.Recorder.Records(TestTopic), 0, false)
	}
}

func SubTestSubscriber(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-topic-subscribe`
		s, e := di.Binder.Subscribe(topic)
		g.Expect(e).To(Succeed(), "Subscribe() should not fail")
		g.Expect(s.Topic()).To(Equal(topic), "subscriber's Topic() should be correct")
		g.Expect(s.Partitions()).ToNot(BeEmpty(), "subscriber's Partitions() should be correct")
		g.Expect(di.Binder.ListTopics()).To(ContainElement(topic), "binder ListTopics() should be correct")
		e = s.AddHandler(di.Service.Handle)
		g.Expect(e).To(Succeed(), "subscriber's AddHandler should not fail")

		// mock some message
		mock := &kafka.Message{
			Payload: &TestMessage{
				Int:    10,
				String: "mya",
			},
		}
		di.Service.received = nil
		e = di.Mocker.Mock(ctx, topic, mock)
		g.Expect(e).To(Succeed(), "mocking incoming message should not fail")
		e = di.Mocker.Mock(ctx, `another topic`, mock)
		g.Expect(e).To(Succeed(), "mocking incoming message should not fail")
		g.Expect(di.Service.received).To(HaveLen(1), "subscriber should received correct message count")
		g.Expect(di.Service.received).To(ContainElement(BeEquivalentTo(mock.Payload)), "subscriber should received correct messages")
	}
}

func SubTestConsumer(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-topic-consume`
		const group = `test-group`
		c, e := di.Binder.Consume(topic, group)
		g.Expect(e).To(Succeed(), "Consume() should not fail")
		g.Expect(c.Topic()).To(Equal(topic), "consumer's Topic() should be correct")
		g.Expect(c.Group()).To(Equal(group), "consumer's Group() should be correct")
		g.Expect(di.Binder.ListTopics()).To(ContainElement(topic), "binder ListTopics() should be correct")
		e = c.AddHandler(di.Service.Handle)
		g.Expect(e).To(Succeed(), "consumer's AddHandler should not fail")

		// mock some message
		mock := &kafka.Message{
			Payload: &TestMessage{
				Int:    10,
				String: "mya",
			},
		}
		di.Service.received = nil
		e = di.Mocker.MockWithGroup(ctx, topic, group, mock)
		g.Expect(e).To(Succeed(), "mocking incoming message should not fail")
		e = di.Mocker.MockWithGroup(ctx, `another topic`, group, mock)
		g.Expect(e).To(Succeed(), "mocking incoming message should not fail")
		e = di.Mocker.MockWithGroup(ctx, topic, "another group", mock)
		g.Expect(e).To(Succeed(), "mocking incoming message should not fail")
		g.Expect(di.Service.received).To(HaveLen(1), "subscriber should received correct message count")
		g.Expect(di.Service.received).To(ContainElement(BeEquivalentTo(mock.Payload)), "subscriber should received correct messages")
	}
}

/*************************
	Helpers
 *************************/

func assertRecordedMessages(_ *testing.T, g *gomega.WithT, actual []*MessageRecord, expectedCount int, verifyMsgs bool) {
	g.Expect(actual).To(HaveLen(expectedCount), "recorded messages should have correct length")
	if !verifyMsgs {
		return
	}

	for i, record := range actual {
		g.Expect(record.Payload).To(BeAssignableToTypeOf(&TestMessage{}), "recorded message at [%d] should have correct type", i)
		msg := record.Payload.(*TestMessage)
		g.Expect(msg.Int).To(BeEquivalentTo(i), "recorded message at [%d] should have correct Int field", i)
		g.Expect(msg.String).To(BeEquivalentTo(fmt.Sprintf("Message-%d", i)), "recorded message at [%d] should have correct String field", i)
	}
}
