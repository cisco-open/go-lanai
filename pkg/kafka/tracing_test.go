package kafka_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/kafka"
	"github.com/cisco-open/go-lanai/pkg/kafka/testdata"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"go.uber.org/fx"
	"strconv"
	"testing"
	"time"
)

const (
	testHeaderTraceId = `mockpfx-ids-traceid`
	testHeaderSpanId = `mockpfx-ids-spanid`
	testHeaderSampled = `mockpfx-ids-sampled`
)

/*************************
	Test Setup
 *************************/

func NewTestTracer() (opentracing.Tracer, *mocktracer.MockTracer) {
	tracer := mocktracer.New()
	return tracer, tracer
}

/*************************
	Tests
 *************************/

type TestTracingDI struct {
	fx.In
	TestBinderDI
	testdata.MockHeadersDI
	MockTracer *mocktracer.MockTracer
}

func TestKafkaTracing(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(60*time.Second),
		testdata.WithMockedBroker(),
		apptest.WithModules(kafka.Module),
		apptest.WithFxOptions(
			fx.Provide(NewTestTracer),
			fx.Provide(testdata.ProvideMockedHeadersInterceptor),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestResetTracer(&di)),
		test.SubTestSetup(testdata.SubSetupHeadersMocker(&di.MockHeadersDI)),
		test.SubTestSetup(SetupTestStartSpan(&di)),
		test.GomegaSubTest(SubTestProducerTracing(&di), "TestProducerTracing"),
		test.GomegaSubTest(SubTestSubscriberTracingWithExistingSpan(&di), "TestSubscriberTracingWithExistingSpan"),
		test.GomegaSubTest(SubTestSubscriberTracingWithoutExistingSpan(&di), "TestSubscriberTracingWithoutExistingSpan"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SetupTestResetTracer(di *TestTracingDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		di.MockTracer.Reset()
		return ctx, nil
	}
}

func SetupTestStartSpan(di *TestTracingDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		ctx, _ = ContextWithTestSpan(ctx, di.MockTracer)
		return ctx, nil
	}
}

func SubTestProducerTracing(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test.producer-tracing`
		var e error
		span := FindSpan(ctx)
		encoder := &TestEncoder{}
		producer := TryBindTestProducer(ctx, t, g, &di.TestBinderDI, topic, kafka.AckTimeout(1*time.Second))
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

		AssertProducerSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "send", topic)
	}
}

func SubTestSubscriberTracingWithExistingSpan(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test.pubsub-tracing`
		var e error
		span := FindSpan(ctx)
		subscriber := TryBindTestSubscriber(ctx, g, &di.TestBinderDI, topic)
		// setup handler
		ch := make(chan *mocktracer.MockSpan, 1)
		defer close(ch)
		e = subscriber.AddHandler(func(ctx context.Context, payload map[string]interface{}) error {
			ch <- FindSpan(ctx)
			return nil
		})
		g.Expect(e).To(Succeed(), "adding handler should not fail")

		// mock some messages and wait for trigger
		payload := map[string]interface{}{"value": "hello"}
		go testdata.MockSubscribedMessage(ctx, topic, 0, 0, MakeMockedMessage(
			WithValue(payload),
			WithHeader(testHeaderTraceId, strconv.Itoa(span.SpanContext.TraceID)),
			WithHeader(testHeaderSpanId, strconv.Itoa(span.SpanContext.SpanID)),
			WithHeader(testHeaderSampled, "true"),
		))
		capturedSpan, e := WaitForHandlerInvocation(ctx, ch, 10*time.Second)
		g.Expect(e).To(Succeed(), "handler 1 should be triggered")
		parent := Last(di.MockTracer.FinishedSpans())
		AssertSubscribeSpan(ctx, g, parent, span, "subscribe", false)
		AssertSubscribeSpan(ctx, g, capturedSpan, parent, "handle", false)
	}
}

func SubTestSubscriberTracingWithoutExistingSpan(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const topic = `test.pubsub-tracing-no-span`
		var e error
		subscriber := TryBindTestSubscriber(ctx, g, &di.TestBinderDI, topic)
		// setup handler
		ch := make(chan *mocktracer.MockSpan, 1)
		defer close(ch)
		e = subscriber.AddHandler(func(ctx context.Context, payload map[string]interface{}) error {
			ch <- FindSpan(ctx)
			return errors.New("oops")
		})
		g.Expect(e).To(Succeed(), "adding handler should not fail")

		// mock some messages and wait for trigger
		payload := map[string]interface{}{"value": "hello"}
		go testdata.MockSubscribedMessage(ctx, topic, 0, 0, MakeMockedMessage(WithValue(payload)))
		capturedSpan, e := WaitForHandlerInvocation(ctx, ch, 10*time.Second)
		g.Expect(e).To(Succeed(), "handler 1 should be triggered")
		parent := Last(di.MockTracer.FinishedSpans())
		AssertSubscribeSpan(ctx, g, parent, nil, "subscribe", true)
		AssertSubscribeSpan(ctx, g, capturedSpan, parent, "handle", true)
	}
}

/*************************
	Helper
 *************************/

func ContextWithTestSpan(ctx context.Context, tracer opentracing.Tracer) (context.Context, *mocktracer.MockSpan) {
	ctx = tracing.WithTracer(tracer).
		WithOpName("test").
		WithOptions(tracing.SpanKind(ext.SpanKindRPCServerEnum)).
		NewSpanOrDescendant(ctx)
	return ctx, FindSpan(ctx)
}

func FindSpan(ctx context.Context) *mocktracer.MockSpan {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		return span.(*mocktracer.MockSpan)
	}
	return nil
}

func Last[T any](slice []T) (last T) {
	if len(slice) == 0 {
		return
	}
	return slice[len(slice)-1]
}

func ExpectedOpName(op string) string {
	return fmt.Sprintf(`kafka %s`, op)
}

func AssertProducerSpans(_ context.Context, g *gomega.WithT, spans []*mocktracer.MockSpan, expectedParent *mocktracer.MockSpan, expectedOp, expectedTopic string) *mocktracer.MockSpan {
	if expectedParent == nil || len(expectedOp) == 0 {
		g.Expect(spans).To(BeEmpty(), "recorded span should be empty")
		return nil
	}
	span := Last(spans)
	g.Expect(span).ToNot(BeNil(), "recorded span should be available")
	g.Expect(span.OperationName).To(Equal(ExpectedOpName(expectedOp)), "recorded span should have correct '%s'", "OpName")
	g.Expect(span.SpanContext.TraceID).To(Equal(expectedParent.SpanContext.TraceID), "recorded span should have correct '%s'", "TraceID")
	g.Expect(span.ParentID).To(Equal(expectedParent.SpanContext.SpanID), "recorded span should have correct '%s'", "ParentID")
	g.Expect(span.Tags()).To(HaveKeyWithValue("span.kind", BeEquivalentTo("client")), "recorded span should have correct '%s'", "Tags")
	g.Expect(span.Tags()).To(HaveKeyWithValue("topic", expectedTopic), "recorded span should have correct '%s'", "Tags")
	g.Expect(span.Tags()).To(HaveKeyWithValue("key", Not(BeZero())), "recorded span should have correct '%s'", "Tags")
	g.Expect(span.Tags()).To(HaveKey("partition"), "recorded span should have correct '%s'", "Tags")
	g.Expect(span.Tags()).To(HaveKey("offset"), "recorded span should have correct '%s'", "Tags")

	return span
}


func AssertSubscribeSpan(_ context.Context, g *gomega.WithT, span *mocktracer.MockSpan, expectedAncestor *mocktracer.MockSpan, expectedOp string, expectErr bool) *mocktracer.MockSpan {
	g.Expect(span).ToNot(BeNil(), "recorded subscribe span should be available")
	g.Expect(span.OperationName).To(Equal(ExpectedOpName(expectedOp)), "recorded span should have correct '%s'", "OpName")
	if expectedAncestor != nil {
		// note: we don't check parent ID because the expected ParentID is not known
		g.Expect(span.SpanContext.TraceID).To(Equal(expectedAncestor.SpanContext.TraceID), "recorded span should have correct '%s'", "TraceID")
	} else {
		g.Expect(span.ParentID).ToNot(BeZero(), "recorded span should have correct '%s' when following", "ParentID")
		g.Expect(span.SpanContext.TraceID).ToNot(Equal(0), "recorded span should have correct '%s'", "TraceID")
	}
	g.Expect(span.Tags()).To(HaveKeyWithValue("span.kind", BeEquivalentTo("server")), "recorded span should have correct '%s'", "Tags")
	if expectErr {
		g.Expect(span.Tags()).To(HaveKeyWithValue("err", Not(BeZero())), "recorded span should have correct '%s'", "Tags")
	}
	return span
}