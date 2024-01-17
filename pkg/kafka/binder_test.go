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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/kafka"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/kafka/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"fmt"
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

type TestBinderDI struct {
	fx.In
	AppContext *bootstrap.ApplicationContext
	Binder     kafka.Binder
}

func TestBinder(t *testing.T) {
	di := TestBinderDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120 * time.Second),
		testdata.WithMockedBroker(),
		apptest.WithFxOptions(
			fx.Provide(kafka.BindKafkaProperties, kafka.ProvideKafkaBinder),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SubSetupStartBinder(&di)),
		test.GomegaSubTest(SubTestBindProducer(&di), "TestBindProducer"),
		test.GomegaSubTest(SubTestBindProducerAddPartition(&di), "TestBindProducerAddPartition"),
		test.GomegaSubTest(SubTestBindSubscriber(&di), "TestBindSubscriber"),
		test.GomegaSubTest(SubTestBindConsumer(&di), "TestBindConsumer"),
		test.GomegaSubTest(SubTestBinderHealth(&di), "TestBinderHealth"),
		test.GomegaSubTest(SubTestShutdown(&di), "TestShutdown"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubSetupStartBinder(di *TestBinderDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		// note: repeatedly restart binder should do nothing
		g := gomega.NewWithT(t)
		g.Expect(di.Binder).To(BeAssignableToTypeOf(kafka.BinderLifecycle(&kafka.SaramaKafkaBinder{})))
		e := di.Binder.(kafka.BinderLifecycle).Start(ctx)
		g.Expect(e).To(Succeed(), "starting binder should not fail")
		return ctx, e
	}
}

func SubTestBindProducer(di *TestBinderDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-producer`

		testdata.MockCreateTopic(ctx, topic)
		producer, e := di.Binder.Produce(topic, kafka.RequireLocalAck())
		g.Expect(e).To(Succeed(), "bind producer should not fail")
		g.Expect(producer).ToNot(BeNil(), "producer should not be nil")
		g.Expect(producer.Topic()).To(Equal(topic), "producer's topic should be correct")

		// check topics
		topics := di.Binder.ListTopics()
		g.Expect(topics).To(ContainElement(topic), "list topics should be correct")
	}
}

func SubTestBindProducerAddPartition(di *TestBinderDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-partition-add`

		testdata.MockExistingTopic(ctx, topic, 0)
		testdata.MockCreatePartition(ctx, topic, 1)
		producer, e := di.Binder.Produce(topic, kafka.RequireLocalAck(), kafka.Partitions(3, 2))
		g.Expect(e).To(Succeed(), "bind producer should not fail")
		g.Expect(producer).ToNot(BeNil(), "producer should not be nil")
		g.Expect(producer.Topic()).To(Equal(topic), "producer's topic should be correct")

		// check topics
		topics := di.Binder.ListTopics()
		g.Expect(topics).To(ContainElement(topic), "list topics should be correct")
		fmt.Println("Bind Producer finished")
	}
}

func SubTestBindSubscriber(di *TestBinderDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-pubsub`

		testdata.MockExistingTopic(ctx, topic, 0)
		subscriber, e := di.Binder.Subscribe(topic)
		g.Expect(e).To(Succeed(), "bind subscriber should not fail")
		g.Expect(subscriber).ToNot(BeNil(), "subscriber should not be nil")
		g.Expect(subscriber.Topic()).To(Equal(topic), "subscriber's topic should be correct")

		topics := di.Binder.ListTopics()
		g.Expect(topics).To(ContainElement(topic), "list topics should be correct")
	}
}

func SubTestBindConsumer(di *TestBinderDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test-consumer`
		const group = `test.group`

		testdata.MockExistingTopic(ctx, topic, 0)
		consumer, e := di.Binder.Consume(topic, group)
		g.Expect(e).To(Succeed(), "bind consumer should not fail")
		g.Expect(consumer).ToNot(BeNil(), "consumer should not be nil")
		g.Expect(consumer.Topic()).To(Equal(topic), "consumer's topic should be correct")
		g.Expect(consumer.Group()).To(Equal(group), "consumer's group should be correct")

		topics := di.Binder.ListTopics()
		g.Expect(topics).To(ContainElement(topic), "list topics should be correct")
	}
}

func SubTestBinderHealth(di *TestBinderDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		indicator := kafka.NewHealthIndicator(di.Binder)
		h := indicator.Health(ctx, health.Options{
			ShowDetails:    true,
			ShowComponents: true,
		})
		g.Expect(h.Status()).To(Equal(health.StatusUp), "status should be correct")
		g.Expect(h.Description()).ToNot(BeEmpty(), "description should be correct")
		g.Expect(h).To(BeAssignableToTypeOf(&health.DetailedHealth{}))
		detailed := h.(*health.DetailedHealth)
		g.Expect(detailed.Details).To(HaveKeyWithValue("topics", HaveLen(4)), "details should be correct")
	}
}

func SubTestShutdown(di *TestBinderDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		syncCh := make(chan struct{}, 1)
		go func(done <-chan struct{}) {
			timeoutCtx, cancelFn := context.WithTimeout(ctx, 1*time.Second)
			defer cancelFn()
			select {
			case <-done:
			case <-timeoutCtx.Done():
				t.Errorf(`BinderLifecycle.Done() is not triggered during shutdown`)
			}
			close(syncCh)
		}(di.Binder.(kafka.BinderLifecycle).Done())

		e := di.Binder.(kafka.BinderLifecycle).Shutdown(ctx)
		g.Expect(e).To(Succeed(), "starting binder should not fail")

		// validate Done() channel after Shutdown
		timeoutCtx, cancelFn := context.WithTimeout(ctx, 1*time.Second)
		defer cancelFn()
		select {
		case <-di.Binder.(kafka.BinderLifecycle).Done():
		case <-timeoutCtx.Done():
			t.Errorf(`BinderLifecycle.Done() is not triggered during shutdown`)
		}

		// wait a bit
		select { case <-syncCh: }

		// validate post-shutdown behavior
		for _, topic := range di.Binder.ListTopics() {
			_, e = di.Binder.Produce(topic)
			g.Expect(e).To(HaveOccurred(), "attempting to bind producer after shutdown should fail")

			// Note use a closed Binder to bind Subscriber or Consumer is currently not prohibited, but it won't receive any messages

			if subscriber, e := di.Binder.Subscribe(topic); e != nil {
				g.Expect(subscriber.(kafka.BindingLifecycle).Closed()).To(BeTrue(), "existing subscriber should be closed")
			}
			if consumer, e := di.Binder.Consume(topic, "test.group"); e != nil {
				g.Expect(consumer.(kafka.BindingLifecycle).Closed()).To(BeTrue(), "existing consumer should be closed")
			}
		}
	}
}

/*************************
	Helpers
 *************************/



