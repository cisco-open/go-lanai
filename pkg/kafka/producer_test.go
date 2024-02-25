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
    "github.com/cisco-open/go-lanai/pkg/kafka"
    "github.com/cisco-open/go-lanai/pkg/kafka/testdata"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/google/uuid"
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

type TestProducerDI struct {
	fx.In
	TestBinderDI
}

func TestProducer(t *testing.T) {
	di := TestProducerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		testdata.WithMockedBroker(),
		apptest.WithModules(kafka.Module),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestSendWithLocalAck(&di), "TestSendWithLocalAck"),
		test.GomegaSubTest(SubTestSendWithoutAck(&di), "TestSendWithoutAck"),
		test.GomegaSubTest(SubTestSendWithAllAck(&di), "TestSendWithAllAck"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestSendWithLocalAck(di *TestProducerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test.producer-local-ack`
		var e error
		encoder := &TestEncoder{}
		producer := TryBindTestProducer(ctx, t, g, di, topic, kafka.RequireLocalAck(), kafka.AckTimeout(1 * time.Second))
		g.Expect(producer.Topic()).To(Equal(topic), "producer's topic should be correct")

		// send some messages
		testdata.MockProduce(ctx, topic, false)
		msg := kafka.Message{
			Headers: kafka.Headers{"test-header": "test-header-value"},
			Payload: map[string]interface{}{
				"test-body-int-field":    1,
				"test-body-string-field": "value",
			},
		}

		e = producer.Send(ctx, msg, kafka.WithKey(uuid.New()), kafka.WithEncoder(encoder))
		g.Expect(e).To(Succeed(), "producer Send(msg) should not fail")
		g.Expect(encoder.MIMETypeCount).To(Equal(1), "encoder.MIMEType should be called")
		g.Expect(encoder.EncodeCount).To(Equal(1), "encoder.Encode should be called")
	}
}

func SubTestSendWithoutAck(di *TestProducerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test.producer-no-ack`
		var e error
		keyEncoder := &TestEncoder{}
		producer := TryBindTestProducer(ctx, t, g, di, topic, kafka.KeyEncoder(keyEncoder), kafka.RequireNoAck())
		g.Expect(producer.Topic()).To(Equal(topic), "producer's topic should be correct")

		// send some messages
		testdata.MockProduce(ctx, topic, false)
		msg := kafka.Message{
			Headers: kafka.Headers{"test-header": "test-header-value"},
			Payload: map[string]interface{}{
				"test-body-int-field":    1,
				"test-body-string-field": "value",
			},
		}
		encoder := &TestEncoder{}
		e = producer.Send(ctx, msg.Payload, kafka.WithKey(uuid.New()), kafka.WithEncoder(encoder))
		g.Expect(e).To(Succeed(), "producer Send(msg.Payload) should not fail")
		g.Expect(encoder.MIMETypeCount).To(Equal(1), "encoder.MIMEType should be called")
		g.Expect(encoder.EncodeCount).To(Equal(1), "encoder.Encode should be called")
		g.Expect(keyEncoder.MIMETypeCount).To(Equal(0), "keyEncoder.MIMEType should not be called")
		g.Expect(keyEncoder.EncodeCount).To(Equal(1), "keyEncoder.Encode should be called")
	}
}

func SubTestSendWithAllAck(di *TestProducerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test.producer-all-ack`
		var e error
		producer := TryBindTestProducer(ctx, t, g, di, topic, kafka.RequireAllAck(), kafka.AckTimeout(1 * time.Second))
		g.Expect(producer.Topic()).To(Equal(topic), "producer's topic should be correct")

		// send some messages
		testdata.MockProduce(ctx, topic, false)
		msg := kafka.Message{
			Headers: kafka.Headers{"test-header": "test-header-value"},
			Payload: map[string]interface{}{
				"test-body-int-field":    1,
				"test-body-string-field": "value",
			},
		}
		e = producer.Send(ctx, &msg)
		g.Expect(e).To(Succeed(), "producer Send(&msg) should not fail")
	}
}

/*************************
	Helpers
 *************************/

func TryBindTestProducer(ctx context.Context, t *testing.T, g *gomega.WithT, di *TestProducerDI, topic string, opts...kafka.ProducerOptions) kafka.Producer {
	testdata.MockCreateTopic(ctx, topic)
	producer, e := di.Binder.Produce(topic, opts...)
	g.Expect(e).To(Succeed(), "bind producer should not fail")
	g.Expect(producer).ToNot(BeNil(), "producer should not be nil")


	// wait for producer to be ready
	timeoutCtx, cancelFn := context.WithTimeout(ctx, 1*time.Second)
	defer cancelFn()
	select {
	case <-producer.ReadyCh():
	case <-timeoutCtx.Done():
		t.Errorf(`producer did not become "ready"`)
	}
	return producer
}

type TestEncoder struct {
	MIMETypeCount int
	EncodeCount   int
}

func (enc *TestEncoder) Reset() {
	enc.MIMETypeCount = 0
	enc.EncodeCount = 0
}

func (enc *TestEncoder) MIMEType() string {
	enc.MIMETypeCount ++
	return "application/json"
}

func (enc *TestEncoder) Encode(v interface{}) ([]byte, error) {
	enc.EncodeCount ++
	return json.Marshal(v)
}

