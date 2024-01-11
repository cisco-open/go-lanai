package kafka_test

import (
	"context"
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

const topic = `test-consumer-dispatch`
const group = `test.group`

func ProvideTestGroupConsumer(binder kafka.Binder, lc fx.Lifecycle) (kafka.GroupConsumer, *TestHandler, error) {
	consumer, e := binder.Consume(topic, group)
	if e != nil {
		return nil, nil, e
	}
	handler := &TestHandler{
		CH: make(chan HandlerParams, 1),
	}
	lc.Append(fx.StopHook(func(context.Context) { close(handler.CH) }))
	return consumer, handler, consumer.AddHandler(handler.HandleFunc)
}

type TestHandler struct {
	CH    chan HandlerParams
	Error error
}

func (h *TestHandler) HandleFunc(_ context.Context, raw *kafka.Message, meta *kafka.MessageMetadata) error {
	h.CH <- HandlerParams{Message: raw, Metadata: meta}
	return h.Error
}

/*************************
	Tests
 *************************/

type TestConsumerDI struct {
	fx.In
	TestBinderDI
	Consumer kafka.GroupConsumer
	Handler  *TestHandler
}

func TestConsumer(t *testing.T) {
	di := TestConsumerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(120*time.Second),
		testdata.WithMockedBroker(),
		apptest.WithModules(kafka.Module),
		apptest.WithFxOptions(
			// Note: It takes time for consumer to join group. We don't want to do it repeatedly
			fx.Provide(ProvideTestGroupConsumer),
			fx.Provide(testdata.ProvideMockedHeadersInterceptor),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SubSetupStartBinder(&di.TestBinderDI)),
		test.GomegaSubTest(SubTestConsumerDispatch(&di), "TestDispatch"),
		test.GomegaSubTest(SubTestConsumerDispatchWithErrorResult(&di), "DispatchWithErrorResult"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestConsumerDispatch(di *TestConsumerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var v HandlerParams
		testdata.MockExistingTopic(ctx, topic, 0)
		testdata.MockExistingTopic(ctx, topic, 1)
		testdata.MockGroup(ctx, topic, group, 0)

		// mock some messages
		go testdata.MockGroupMessage(ctx, topic, group, 0, 0, MakeMockedMessage(WithValue([]byte("binary"))))
		v, e = WaitForHandlerInvocation(ctx, di.Handler.CH, 50*time.Second)
		g.Expect(e).To(Succeed(), "handler should be triggered")
		g.Expect(v.Message).ToNot(BeNil(), "handler params should have message")
		g.Expect(v.Message.Payload).To(BeEquivalentTo([]byte("binary")), "payload should be correct")
		AssertMetadata(g, v.Metadata, 0, 0, nil)

		// 2nd message
		go testdata.MockGroupMessage(ctx, topic, group, 0, 1, MakeMockedMessage(WithValue([]byte("binary"))))
		v, e = WaitForHandlerInvocation(ctx, di.Handler.CH, 50*time.Second)
		g.Expect(e).To(Succeed(), "handler should be triggered")
		g.Expect(v.Message).ToNot(BeNil(), "handler params should have message")
		g.Expect(v.Message.Payload).To(BeEquivalentTo([]byte("binary")), "payload should be correct")
		AssertMetadata(g, v.Metadata, 0, 1, nil)

		// message with different partition
		go testdata.MockGroupMessage(ctx, topic, group, 1, 0, MakeMockedMessage(WithValue([]byte("not-related"))))
		v, e = WaitForHandlerInvocation(ctx, di.Handler.CH, 200*time.Millisecond)
		g.Expect(e).To(HaveOccurred(), "handler should not be triggered with different partition")
	}
}

func SubTestConsumerDispatchWithErrorResult(di *TestConsumerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var v HandlerParams
		testdata.MockExistingTopic(ctx, topic, 0)
		testdata.MockExistingTopic(ctx, topic, 1)
		testdata.MockGroup(ctx, topic, group, 0)

		// set handler behavior
		di.Handler.Error = fmt.Errorf("oops")

		// mock some messages
		go testdata.MockGroupMessage(ctx, topic, group, 0, 2, MakeMockedMessage(WithValue([]byte("binary"))))
		v, e = WaitForHandlerInvocation(ctx, di.Handler.CH, 5*time.Second)
		g.Expect(e).To(Succeed(), "handler should be triggered")
		g.Expect(v.Message).ToNot(BeNil(), "handler params should have message")
		g.Expect(v.Message.Payload).To(BeEquivalentTo([]byte("binary")), "payload should be correct")
		AssertMetadata(g, v.Metadata, 0, 2, nil)

		// try again
		di.Handler.Error = nil
		// TODO this test case is intended to test if Offset is reset after failure and message is replayed later.
		// 		However, due to our current mocking approach, we cannot dynamically update OffsetResponse in mocked broker
		// 		based on OffsetCommitRequest. Therefore, we cannot expect the handler to be triggered again.
		//v, e = WaitForHandlerInvocation(ctx, di.Handler.CH, 5*time.Second)
		//g.Expect(e).To(Succeed(), "handler should be re-triggered")
		//AssertMetadata(g, v.Metadata, 0, 2, nil)
	}
}

/*************************
	Helpers
 *************************/
