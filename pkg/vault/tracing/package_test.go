package vaulttracing_test

import (
	"context"
	"embed"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/pkg/vault"
	vaultinit "github.com/cisco-open/go-lanai/pkg/vault/init"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/ittest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"testing"
)

//go:embed testdata/bootstrap-test.yml
var TestBootstrapFS embed.FS

/*************************
	Test Setup
 *************************/

func RecordedVaultProvider() fx.Annotated {
	return fx.Annotated{
		Group: "vault",
		Target: func(recorder *recorder.Recorder) vault.Options {
			return func(cfg *vault.ClientConfig) error {
				recorder.SetRealTransport(cfg.HttpClient.Transport)
				cfg.HttpClient.Transport = recorder
				return nil
			}
		},
	}
}

func NewTestTracer() (opentracing.Tracer, *mocktracer.MockTracer) {
	tracer := mocktracer.New()
	return tracer, tracer
}

/*************************
	Tests
 *************************/

type TestTracingDI struct {
	fx.In
	Vault      *vault.Client
	MockTracer *mocktracer.MockTracer
}

func TestAppConfig(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		ittest.WithHttpPlayback(t,
			//ittest.HttpRecordingMode(),
			ittest.DisableHttpRecordOrdering(),
		),
		apptest.WithBootstrapConfigFS(TestBootstrapFS),
		apptest.WithModules(vaultinit.Module),
		apptest.WithFxOptions(
			fx.Provide(RecordedVaultProvider(), NewTestTracer),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestResetHook(&di)),
		test.GomegaSubTest(SubTestReadWithExistingSpan(&di)),
		test.GomegaSubTest(SubTestWriteWithExistingSpan(&di)),
		test.GomegaSubTest(SubTestReadWithoutExistingSpan(&di)),
		test.GomegaSubTest(SubTestWriteWithoutExistingSpan(&di)),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SetupTestResetHook(di *TestTracingDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		di.MockTracer.Reset()
		return ctx, nil
	}
}

func SubTestReadWithExistingSpan(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const path = "secret/test/tracing"
		var e error
		ctx, span := ContextWithTestSpan(ctx, di.MockTracer)
		_, e = di.Vault.Logical(ctx).Read(path)
		AssertSpans(g, di.MockTracer.FinishedSpans(), span, ExpectedOpName("Read", path), e)

		_, e = di.Vault.Logical(ctx).ReadWithData(path, map[string][]string{})
		AssertSpans(g, di.MockTracer.FinishedSpans(), span, ExpectedOpName("Read", path), e)
	}
}

func SubTestWriteWithExistingSpan(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const path = "secret/test/tracing"
		var e error
		ctx, span := ContextWithTestSpan(ctx, di.MockTracer)
		_, e = di.Vault.Logical(ctx).Write(path, map[string]interface{}{
			"test-key": "test-value",
		})
		AssertSpans(g, di.MockTracer.FinishedSpans(), span, ExpectedOpName("Write", path), e)

		_, e = di.Vault.Logical(ctx).Post(path, map[string]interface{}{
			"test-key": "test-value",
		})
		AssertSpans(g, di.MockTracer.FinishedSpans(), span, ExpectedOpName("Post", path), e)

		_, e = di.Vault.Logical(ctx).WriteWithMethod("Put", path, map[string]interface{}{
			"test-key": "test-value",
		})
		AssertSpans(g, di.MockTracer.FinishedSpans(), span, ExpectedOpName("Put", path), e)

		// error case
		_, e = di.Vault.Logical(ctx).WriteWithMethod("Get", path, map[string]interface{}{
			"test-key": "test-value",
		})
		AssertSpans(g, di.MockTracer.FinishedSpans(), span, ExpectedOpName("Get", path), e)
	}
}

func SubTestReadWithoutExistingSpan(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const path = "secret/test/tracing"
		var e error
		_, e = di.Vault.Logical(ctx).Read(path)
		AssertSpans(g, di.MockTracer.FinishedSpans(), nil, "", e)

		_, e = di.Vault.Logical(ctx).ReadWithData(path, map[string][]string{})
		AssertSpans(g, di.MockTracer.FinishedSpans(), nil, "", e)
	}
}

func SubTestWriteWithoutExistingSpan(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const path = "secret/test/tracing"
		var e error
		_, e = di.Vault.Logical(ctx).Write(path, map[string]interface{}{
			"test-key": "test-value",
		})
		AssertSpans(g, di.MockTracer.FinishedSpans(), nil, "", e)

		_, e = di.Vault.Logical(ctx).Post(path, map[string]interface{}{
			"test-key": "test-value",
		})
		AssertSpans(g, di.MockTracer.FinishedSpans(), nil, "", e)

		_, e = di.Vault.Logical(ctx).WriteWithMethod("Put", path, map[string]interface{}{
			"test-key": "test-value",
		})
		AssertSpans(g, di.MockTracer.FinishedSpans(), nil, "", e)

		// error case
		_, e = di.Vault.Logical(ctx).WriteWithMethod("Get", path, map[string]interface{}{
			"test-key": "test-value",
		})
		AssertSpans(g, di.MockTracer.FinishedSpans(), nil, "", e)
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

func ExpectedOpName(op, path string) string {
	return fmt.Sprintf(`vault %s %s`, op, path)
}

func AssertSpans(g *gomega.WithT, spans []*mocktracer.MockSpan, expectedParent *mocktracer.MockSpan, expectedOp string, expectedErr error) {
	if expectedParent == nil || len(expectedOp) == 0 {
		g.Expect(spans).To(BeEmpty(), "recorded span should be empty")
		return
	}
	span := Last(spans)
	g.Expect(span).ToNot(BeNil(), "recorded span should be available")
	g.Expect(span.OperationName).To(Equal(expectedOp), "recorded span should have correct '%s'", "OpName")
	g.Expect(span.SpanContext.TraceID).To(Equal(expectedParent.SpanContext.TraceID), "recorded span should have correct '%s'", "TraceID")
	g.Expect(span.ParentID).To(Equal(expectedParent.SpanContext.SpanID), "recorded span should have correct '%s'", "ParentID")
	if expectedErr != nil {
		g.Expect(span.Tags()).To(HaveKeyWithValue("err", expectedErr), "recorded span should have correct '%s'", "err tag")
	} else {
		g.Expect(span.Tags()).ToNot(HaveKey("err"), "recorded span should have correct '%s'", "err tag")
	}
}