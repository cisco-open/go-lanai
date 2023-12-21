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
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"fmt"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
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
}

func NewTestService(di tsDI) *TestService {
	svc := &TestService{}
	p, e := di.Binder.Produce(TestTopic, kafka.BindingName("test"), kafka.RequireLocalAck())
	if e != nil {
		panic(e)
	}
	svc.producer = p

	s, e := di.Binder.Subscribe(TestTopic)
	if e != nil {
		panic(e)
	}
	if e := s.AddHandler(svc.Handle); e != nil {
		panic(e)
	}

	c, e := di.Binder.Consume(TestTopic, TestGroup)
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

func (s *TestService) Handle(_ context.Context, _ *kafka.Message) error {
	//noop
	return nil
}

/*************************
	Tests
 *************************/

type testDI struct {
	fx.In
	Recorder MessageRecorder `optional:"true"`
	Service  *TestService    `optional:"true"`
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
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestProducerRecording(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Recorder).NotTo(BeNil(), "MessageRecorder should be available")
		g.Expect(di.Service).NotTo(BeNil(), "TestService should be available")

		var e error
		di.Recorder.Reset()
		e = di.Service.GenerateSomeMessages(ctx, 3)
		g.Expect(e).To(Succeed(), "functions using producers shouldn't fail")
		assertRecordedMessages(t, g, di.Recorder.Records(TestTopic), 3, true)

		// do it again without reset
		e = di.Service.GenerateSomeMessages(ctx, 3)
		g.Expect(e).To(Succeed(), "functions using producers shouldn't fail")
		assertRecordedMessages(t, g, di.Recorder.Records(TestTopic), 6, false)

		// validate reset
		di.Recorder.Reset()
		assertRecordedMessages(t, g, di.Recorder.Records(TestTopic), 0, false)
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
