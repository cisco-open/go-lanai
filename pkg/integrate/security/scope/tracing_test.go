package scope_test

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope/testdata"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
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
	return tracer, tracer
}

/*************************
	Tests
 *************************/

type TestTracingDI struct {
	fx.In
	MockTracer *mocktracer.MockTracer
}

func TestHttpClientTracing(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(scope.Module),
		sectest.WithMockedScopes(testdata.TestAcctsFS, testdata.TestBasicFS),
		apptest.WithFxOptions(
			fx.Provide(NewTestTracer),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestResetTracer(&di)),
		test.GomegaSubTest(SubTestTraceWithExistingSpan(&di), "TestTraceWithExistingSpan"),
		test.GomegaSubTest(SubTestTraceWithoutExistingSpan(&di), "TestTraceWithoutExistingSpan"),
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

func SubTestTraceWithExistingSpan(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		ctx, span := ContextWithTestSpan(ctx, di.MockTracer)
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(securityMockRegular()))
		ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
		var scopeDesc string
		e = scope.Do(ctx, func(ctx context.Context) {
			scopeDesc = scope.Describe(ctx)
		}, scope.UseSystemAccount())
		g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, scopeDesc)
	}
}

func SubTestTraceWithoutExistingSpan(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(securityMockRegular()))
		ctx = ContextWithMockedSecurity(ctx, securityMockRegular())
		var scopeDesc string
		e = scope.Do(ctx, func(ctx context.Context) {
			scopeDesc = scope.Describe(ctx)
		}, scope.UseSystemAccount())
		g.Expect(e).To(Succeed(), "scope manager shouldn't returns error")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), nil, scopeDesc)
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
	return opentracing.SpanFromContext(ctx).(*mocktracer.MockSpan)
}

func Last[T any](slice []T) (last T) {
	if len(slice) == 0 {
		return
	}
	return slice[len(slice)-1]
}

func ExpectedOpName() string {
	return fmt.Sprintf(`security`)
}

func AssertSpans(_ context.Context, g *gomega.WithT, spans []*mocktracer.MockSpan, expectedParent *mocktracer.MockSpan, expectedScope string) *mocktracer.MockSpan {
	if expectedParent == nil || len(expectedScope) == 0 {
		g.Expect(spans).To(BeEmpty(), "recorded span should be empty")
		return nil
	}
	span := Last(spans)
	g.Expect(span).ToNot(BeNil(), "recorded span should be available")
	g.Expect(span.OperationName).To(Equal(ExpectedOpName()), "recorded span should have correct '%s'", "OpName")
	g.Expect(span.SpanContext.TraceID).To(Equal(expectedParent.SpanContext.TraceID), "recorded span should have correct '%s'", "TraceID")
	g.Expect(span.ParentID).To(Equal(expectedParent.SpanContext.SpanID), "recorded span should have correct '%s'", "ParentID")
	g.Expect(span.Tags()).To(HaveKeyWithValue("span.kind", BeEquivalentTo("server")), "recorded span should have correct '%s'", "Tags")
	g.Expect(span.Tags()).To(HaveKeyWithValue("sec.scope", expectedScope), "recorded span should have correct '%s'", "Tags")
	return span
}
