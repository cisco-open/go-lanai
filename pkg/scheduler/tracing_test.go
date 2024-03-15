package scheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Test Setup
 *************************/

func NewTestTracer() (opentracing.Tracer, *mocktracer.MockTracer) {
	tracer := mocktracer.New()
	EnableTracing(tracer)
	return tracer, tracer
}

/*************************
	Tests
 *************************/

type TestTracingDI struct {
	fx.In
	MockTracer    *mocktracer.MockTracer
}

func TestSchedulerTracing(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithFxOptions(
			fx.Provide(NewTestTracer),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestResetTracer(&di)),
		test.SubTestSetup(SetupTestStartSpan(&di)),
		test.GomegaSubTest(SubTestStartNow(&di), "TestStartNow"),
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

func SubTestStartNow(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// prepare task
		const taskName = "test-tracing"
		ch := make(chan *mocktracer.MockSpan, 1)
		defer close(ch)
		tf := func(ctx context.Context) error {
			ch <- FindSpan(ctx)
			return errors.New("oops")
		}

		// run task and verify
		span := FindSpan(ctx)
		canceller, e := RunOnce(tf, Name(taskName))
		g.Expect(e).To(Succeed(), "new task shouldn't return error")
		defer canceller.Cancel()

		// wait and verify
		var capturedSpan *mocktracer.MockSpan
		select {
		case capturedSpan = <-ch:
		case <-canceller.Cancelled():
			t.Errorf("task should not be cancelled")
		case <-ctx.Done():
			t.Errorf("task should be triggered before timeout")
		}
		finishedSpan := Last(di.MockTracer.FinishedSpans())
		AssertSpans(ctx, g, finishedSpan, span, "scheduler", taskName, true)
		AssertSpans(ctx, g, capturedSpan, span, "scheduler", taskName, true)
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
	return fmt.Sprintf(`%s`, op)
}

func AssertSpans(_ context.Context, g *gomega.WithT, span *mocktracer.MockSpan, expectedParent *mocktracer.MockSpan, expectedOp, expectedName string, expectErr bool) *mocktracer.MockSpan {

	g.Expect(span).ToNot(BeNil(), "recorded span should be available")
	g.Expect(span.OperationName).To(Equal(ExpectedOpName(expectedOp)), "recorded span should have correct '%s'", "OpName")
	g.Expect(span.SpanContext.TraceID).ToNot(BeZero(), "recorded span should have correct '%s'", "TraceID")
	g.Expect(span.ParentID).To(BeZero(), "recorded span should have correct '%s'", "ParentID")
	if expectedParent != nil {
		g.Expect(span.SpanContext.TraceID).ToNot(Equal(expectedParent.SpanContext.TraceID), "recorded span should have different '%s'", "TraceID")
	}
	g.Expect(span.Tags()).To(HaveKeyWithValue("span.kind", BeEquivalentTo("client")), "recorded span should have correct '%s'", "Tags")
	g.Expect(span.Tags()).To(HaveKeyWithValue("task", HavePrefix(expectedName+"-")), "recorded span should have correct '%s'", "Tags")

	if expectErr {
		g.Expect(span.Tags()).To(HaveKeyWithValue("err", Not(BeZero())), "recorded span should have correct '%s'", "Tags")
	}
	return span
}
