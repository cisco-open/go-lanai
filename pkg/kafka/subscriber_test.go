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

package kafka_test

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/kafka"
    "github.com/cisco-open/go-lanai/pkg/kafka/testdata"
    "github.com/cisco-open/go-lanai/pkg/utils/matcher"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "testing"
    "time"
)

/*************************
	Setup Test
 *************************/

/*************************
	Tests
 *************************/

type TestSubscriberDI struct {
	fx.In
	TestBinderDI
	testdata.MockHeadersDI
}

func TestSubscriber(t *testing.T) {
	di := TestSubscriberDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(60*time.Second),
		testdata.WithMockedBroker(),
		apptest.WithModules(kafka.Module),
		apptest.WithFxOptions(
			fx.Provide(testdata.ProvideMockedHeadersInterceptor),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(testdata.SubSetupHeadersMocker(&di.MockHeadersDI)),
		test.GomegaSubTest(SubTestSubscriberDispatchWithRawMessage(&di), "DispatchWithRawMessage"),
		test.GomegaSubTest(SubTestSubscriberDispatchWithMetadata(&di), "DispatchWithMetadata"),
		test.GomegaSubTest(SubTestSubscriberDispatchWithHeaders(&di), "DispatchWithHeaders"),
		test.GomegaSubTest(SubTestSubscriberDispatchWithErrorResult(&di), "DispatchWithErrorResult"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestSubscriberDispatchWithRawMessage(di *TestSubscriberDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-pubsub-raw`
		var e error
		var v HandlerParams
		subscriber := TryBindTestSubscriber(ctx, g, &di.TestBinderDI, topic)
		// add handler
		ch := make(chan HandlerParams, 1)
		defer close(ch)
		e = subscriber.AddHandler(func(ctx context.Context, raw *kafka.Message) error {
			ch <- HandlerParams{Message: raw}
			return nil
		})
		g.Expect(e).To(Succeed(), "adding handler should not fail")

		// mock some messages and wait for trigger
		go testdata.MockSubscribedMessage(ctx, topic, 0, 0, MakeMockedMessage(WithValue("hello")))
		v, e = WaitForHandlerInvocation(ctx, ch, 10*time.Second)
		g.Expect(e).To(Succeed(), "handler should be triggered")
		g.Expect(v.Message).ToNot(BeNil(), "handler params should have message")
		g.Expect(v.Message.Payload).To(BeEquivalentTo("hello"), "payload should be correct")

		go testdata.MockSubscribedMessage(ctx, topic, 1, 0, MakeMockedMessage(WithValue("hello")))
		v, e = WaitForHandlerInvocation(ctx, ch, 10*time.Second)
		g.Expect(e).To(Succeed(), "handler should be triggered")
		g.Expect(v.Message).ToNot(BeNil(), "handler params should have message")
		g.Expect(v.Message.Payload).To(BeEquivalentTo("hello"), "payload should be correct")

		// check status (partitions only become available after subscriber is started
		g.Expect(subscriber.Topic()).To(Equal(topic), "subscriber's topic should be correct")
		g.Expect(subscriber.Partitions()).To(HaveLen(2), "subscriber's partitions should be correct")
	}
}

func SubTestSubscriberDispatchWithMetadata(di *TestSubscriberDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-pubsub-meta`
		var e error
		var v HandlerParams
		subscriber := TryBindTestSubscriber(ctx, g, &di.TestBinderDI, topic)
		// add handler
		ch := make(chan HandlerParams, 1)
		defer close(ch)
		e = subscriber.AddHandler(func(ctx context.Context, payload *Payload, meta *kafka.MessageMetadata) error {
			ch <- HandlerParams{Payload: payload, Metadata: meta}
			return nil
		})
		g.Expect(e).To(Succeed(), "adding handler should not fail")

		payload := map[string]interface{}{"value": "hello"}
		// mock some messages and wait for trigger
		go testdata.MockSubscribedMessage(ctx, topic, 0, 0, MakeMockedMessage(WithValue(payload)))
		v, e = WaitForHandlerInvocation(ctx, ch, 10*time.Second)
		g.Expect(e).To(Succeed(), "handler should be triggered")
		g.Expect(v.Payload).To(BeAssignableToTypeOf(&Payload{}), "handler params should have payload")
		AssertMetadata(g, v.Metadata, 0, 0, nil)
		AssertPayload(g, v.Payload, &Payload{}, "hello")

		go testdata.MockSubscribedMessage(ctx, topic, 1, 0, MakeMockedMessage(WithValue(payload), WithKey("test-key")))
		v, e = WaitForHandlerInvocation(ctx, ch, 10*time.Second)
		g.Expect(e).To(Succeed(), "handler should be triggered")
		AssertMetadata(g, v.Metadata, 1, 0, []byte("test-key"))
		AssertPayload(g, v.Payload, &Payload{}, "hello")
	}
}

func SubTestSubscriberDispatchWithHeaders(di *TestSubscriberDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-pubsub-header`
		const headerKey = `x-header`
		var e error
		var v HandlerParams
		subscriber := TryBindTestSubscriber(ctx, g, &di.TestBinderDI, topic)

		// add two handlers with different header filter
		ch1 := make(chan HandlerParams, 1)
		defer close(ch1)
		e = subscriber.AddHandler(func(ctx context.Context, payload map[string]interface{}, headers kafka.Headers) error {
			ch1 <- HandlerParams{Payload: payload, Headers: headers}
			return nil
		}, kafka.FilterOnHeader(headerKey, matcher.WithSubString("-1", false)))
		g.Expect(e).To(Succeed(), "adding handler should not fail")

		ch2 := make(chan HandlerParams, 1)
		defer close(ch2)
		e = subscriber.AddHandler(func(ctx context.Context, payload Payload, headers kafka.Headers) error {
			ch2 <- HandlerParams{Payload: payload, Headers: headers}
			return nil
		}, kafka.FilterOnHeader(headerKey, matcher.WithSubString("-2", false)))
		g.Expect(e).To(Succeed(), "adding handler should not fail")

		payload := map[string]interface{}{"value": "hello"}
		// mock some messages and wait for trigger
		go testdata.MockSubscribedMessage(ctx, topic, 0, 0, MakeMockedMessage(
			WithValue(payload), WithHeader(headerKey, "handler-1"),
		))
		v, e = WaitForHandlerInvocation(ctx, ch1, 10*time.Second)
		g.Expect(e).To(Succeed(), "handler 1 should be triggered")
		AssertPayload(g, v.Payload, map[string]interface{}{}, "hello")
		AssertHeaders(g, v.Headers, headerKey, "handler-1")

		go testdata.MockSubscribedMessage(ctx, topic, 0, 1, MakeMockedMessage(
			WithValue(payload), WithHeader(headerKey, "handler-2"),
		))
		v, e = WaitForHandlerInvocation(ctx, ch2, 10*time.Second)
		g.Expect(e).To(Succeed(), "handler 2 should be triggered")
		AssertPayload(g, v.Payload, Payload{}, "hello")
		AssertHeaders(g, v.Headers, headerKey, "handler-2")
	}
}

func SubTestSubscriberDispatchWithErrorResult(di *TestSubscriberDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-pubsub-error`
		var e error
		var v HandlerParams
		subscriber := TryBindTestSubscriber(ctx, g, &di.TestBinderDI, topic)
		// add handler
		ch := make(chan HandlerParams, 1)
		defer close(ch)
		e = subscriber.AddHandler(func(ctx context.Context, raw *kafka.Message) error {
			ch <- HandlerParams{Message: raw}
			return fmt.Errorf("oops")
		})
		g.Expect(e).To(Succeed(), "adding handler should not fail")

		// mock some messages and wait for trigger
		go testdata.MockSubscribedMessage(ctx, topic, 0, 0, MakeMockedMessage(WithValue("hello")))
		v, e = WaitForHandlerInvocation(ctx, ch, 10*time.Second)
		g.Expect(e).To(Succeed(), "handler should be triggered")
		g.Expect(v.Message).ToNot(BeNil(), "handler params should have message")
		g.Expect(v.Message.Payload).To(BeEquivalentTo("hello"), "payload should be correct")

		// verify handler is not invoked twice
		v, e = WaitForHandlerInvocation(ctx, ch, 200*time.Millisecond)
		g.Expect(e).To(HaveOccurred(), "handler should not be triggered twice due to error")
	}
}

/*************************
	Helpers
 *************************/

func MakeMockedMessage(opts ...func(msg *testdata.MockedMessage)) testdata.MockedMessage {
	msg := testdata.MockedMessage{
		Headers: make(map[string]string),
	}
	for _, fn := range opts {
		fn(&msg)
	}
	return msg
}

func WithValue(value interface{}) func(message *testdata.MockedMessage) {
	return func(msg *testdata.MockedMessage) {
		switch v := value.(type) {
		case string:
			msg.Value = []byte(v)
			msg.Headers[kafka.HeaderContentType] = kafka.MIMETypeText
		case []byte:
			msg.Value = v
			msg.Headers[kafka.HeaderContentType] = kafka.MIMETypeBinary
		default:
			msg.Value, _ = json.Marshal(value)
			msg.Headers[kafka.HeaderContentType] = kafka.MIMETypeJson
		}
	}
}

func WithKey(key string) func(message *testdata.MockedMessage) {
	return func(msg *testdata.MockedMessage) {
		msg.Key = []byte(key)
	}
}

func WithHeader(k, v string) func(message *testdata.MockedMessage) {
	return func(msg *testdata.MockedMessage) {
		msg.Headers[k] = v
	}
}

func AssertPayload[T any](g *gomega.WithT, payload interface{}, expectedType T, expectedValue string) {
	g.Expect(payload).To(BeAssignableToTypeOf(expectedType), "payload should have correct type")
	switch v := payload.(type) {
	case *Payload:
		g.Expect(v.Value).To(Equal(expectedValue), "payload value should be correct")
	case Payload:
		g.Expect(v.Value).To(Equal(expectedValue), "payload value should be correct")
	case map[string]interface{}:
		g.Expect(v).To(HaveKeyWithValue("value", expectedValue), "payload value should be correct")
	default:
		g.Expect(payload).To(BeEquivalentTo(expectedValue), "payload value should be correct")
	}
}

func AssertMetadata(g *gomega.WithT, meta *kafka.MessageMetadata, expectedPartition, expectedOffset int, expectedKey []byte) {
	g.Expect(meta).ToNot(BeNil(), "metadata should not be nil")
	g.Expect(meta.Partition).To(BeEquivalentTo(expectedPartition), "partition should be correct")
	g.Expect(meta.Offset).To(BeEquivalentTo(expectedOffset), "offset should be correct")
	if len(expectedKey) != 0 {
		g.Expect(meta.Key).To(BeEquivalentTo(expectedKey), "partition should be correct")
	}
}

func AssertHeaders(g *gomega.WithT, headers kafka.Headers, expectedKVs ...string) {
	g.Expect(headers).ToNot(BeNil(), "headers should not be nil")
	g.Expect(len(expectedKVs) % 2).To(BeZero(), "expectedKVs should have even length")
	for i := range expectedKVs {
		if i % 2 != 0 {
			continue
		}
		g.Expect(headers).To(HaveKeyWithValue(expectedKVs[i], expectedKVs[i+1]), "headers should contains correct KV")
	}
}

func TryBindTestSubscriber(ctx context.Context, g *gomega.WithT, di *TestBinderDI, topic string, opts ...kafka.ConsumerOptions) kafka.Subscriber {
	testdata.MockExistingTopic(ctx, topic, 0)
	testdata.MockExistingTopic(ctx, topic, 1)
	subscriber, e := di.Binder.Subscribe(topic, opts...)
	g.Expect(e).To(Succeed(), "bind subscriber should not fail")
	g.Expect(subscriber).ToNot(BeNil(), "subscriber should not be nil")
	return subscriber
}

func WaitForHandlerInvocation[T any](ctx context.Context, ch chan T, timeout time.Duration) (T, error) {
	timeoutCtx, cancelFn := context.WithTimeout(ctx, timeout)
	defer cancelFn()
	select {
	case v := <-ch:
		return v, nil
	case <-timeoutCtx.Done():
		var zero T
		return zero, context.DeadlineExceeded
	}
}

type HandlerParams struct {
	Payload  interface{}
	Metadata *kafka.MessageMetadata
	Headers  kafka.Headers
	Message  *kafka.Message
}

type Payload struct {
	Value string `json:"value"`
}
