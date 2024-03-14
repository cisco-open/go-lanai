package redis_test

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/embedded"
	goRedis "github.com/go-redis/redis/v8"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"go.uber.org/fx"
	"testing"
	"time"
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
	ClientFactory   redis.ClientFactory
	MockTracer *mocktracer.MockTracer
}

func TestRedisTracingWithExistingSpan(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		embedded.WithRedis(),
		apptest.WithModules(redis.Module),
		apptest.WithFxOptions(
			fx.Provide(NewTestTracer),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestResetTracer(&di)),
		test.SubTestSetup(SetupTestStartSpan(&di)),
		test.GomegaSubTest(SubTestPing(&di), "TestPing"),
		test.GomegaSubTest(SubTestGet(&di), "TestGet"),
		test.GomegaSubTest(SubTestPipeline(&di), "TestPipeline"),
	)
}

func TestRedisTracingWithoutExistingSpan(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		embedded.WithRedis(),
		apptest.WithModules(redis.Module),
		apptest.WithFxOptions(
			fx.Provide(NewTestTracer),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestResetTracer(&di)),
		test.GomegaSubTest(SubTestPing(&di), "TestPing"),
		test.GomegaSubTest(SubTestGet(&di), "TestGet"),
		test.GomegaSubTest(SubTestPipeline(&di), "TestPipeline"),
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

func SubTestPing(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		span := FindSpan(ctx)
		client, e := di.ClientFactory.New(ctx, func(opt *redis.ClientOption) {
			opt.DbIndex = 5
		})

		g.Expect(e).To(Succeed(), "injected client factory shouldn't return error")
		cmd := client.Ping(ctx)
		g.Expect(cmd).To(Not(BeNil()), "redis ping shouldn't return nil")
		g.Expect(cmd.Err()).To(Succeed(), "redis ping shouldn't return error")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "redis ping", 5,false)
	}
}

func SubTestGet(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		span := FindSpan(ctx)
		client, e := di.ClientFactory.New(ctx, func(opt *redis.ClientOption) {
			opt.DbIndex = 5
		})

		g.Expect(e).To(Succeed(), "injected client factory shouldn't return error")
		cmd := client.Get(ctx, "non-exist")
		g.Expect(cmd).To(Not(BeNil()), "redis GET shouldn't return nil")
		g.Expect(cmd.Err()).To(HaveOccurred(), "redis GET should return error")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "redis get", 5,false)
	}
}

func SubTestPipeline(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		span := FindSpan(ctx)
		client, e := di.ClientFactory.New(ctx, func(opt *redis.ClientOption) {
			opt.DbIndex = 5
		})

		g.Expect(e).To(Succeed(), "injected client factory shouldn't return error")
		cmds, e := client.Pipelined(ctx, func(pipeliner goRedis.Pipeliner) error {
			return pipeliner.Set(ctx, "test-tracing", "some value", time.Second).Err()
		})
		g.Expect(e).To(Succeed(), "redis pipeline should not fail")
		g.Expect(cmds).To(Not(BeEmpty()), "redis pipeline result should not be empty")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "redis-batch", 5,false)
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

func AssertSpans(_ context.Context, g *gomega.WithT, spans []*mocktracer.MockSpan, expectedParent *mocktracer.MockSpan, expectedOp string, expectedDB int, expectErr bool) *mocktracer.MockSpan {
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
	g.Expect(span.Tags()).To(HaveKeyWithValue("cmd", Not(BeZero())), "recorded span should have correct '%s'", "Tags")
	if expectedOp != "redis-batch" {
		g.Expect(span.Tags()).To(HaveKeyWithValue("db", BeEquivalentTo(expectedDB)), "recorded span should have correct '%s'", "Tags")
	}
	if expectErr {
		g.Expect(span.Tags()).To(HaveKeyWithValue("err", Not(BeZero())), "recorded span should have correct '%s'", "Tags")
	}
	return span
}
