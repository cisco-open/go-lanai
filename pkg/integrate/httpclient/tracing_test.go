package httpclient_test

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/integrate/httpclient"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"go.uber.org/fx"
	"net/http"
	"strconv"
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
	HttpClient       httpclient.Client
	MockedController *MockedController
	MockTracer *mocktracer.MockTracer
}

func TestHttpClientTracing(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		apptest.WithModules(httpclient.Module),
		apptest.WithFxOptions(
			fx.Provide(NewMockedController, NewTestTracer),
			web.FxControllerProviders(ProvideWebController),
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
		baseUrl := fmt.Sprintf(`http://localhost:%d%s`, webtest.CurrentPort(ctx), webtest.CurrentContextPath(ctx))
		client, e := di.HttpClient.WithBaseUrl(baseUrl)
		g.Expect(e).To(Succeed(), "client with base URL should be available")

		ctx, span := ContextWithTestSpan(ctx, di.MockTracer)
		reqBody := makeEchoRequestBody()
		req := httpclient.NewRequest(TestPath, http.MethodPost, httpclient.WithBody(reqBody))

		resp, e := client.Execute(ctx, req, httpclient.JsonBody(&EchoResponse{}))
		g.Expect(e).To(Succeed(), "execute request shouldn't fail")
		newSpan := AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, http.MethodPost, http.StatusOK)

		expected := EchoResponse{
			Headers: map[string]string{
				"Mockpfx-Ids-Sampled": "true",
				"Mockpfx-Ids-Spanid": strconv.Itoa(newSpan.SpanContext.SpanID),
				"Mockpfx-Ids-Traceid": strconv.Itoa(newSpan.SpanContext.TraceID),
			},
			ReqBody: reqBody,
		}
		assertResponse(t, g, resp, http.StatusOK, &expected)
	}
}

func SubTestTraceWithoutExistingSpan(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		baseUrl := fmt.Sprintf(`http://localhost:%d%s`, webtest.CurrentPort(ctx), webtest.CurrentContextPath(ctx))
		client, e := di.HttpClient.WithBaseUrl(baseUrl)
		g.Expect(e).To(Succeed(), "client with base URL should be available")

		reqBody := makeEchoRequestBody()
		req := httpclient.NewRequest(TestPath, http.MethodPost, httpclient.WithBody(reqBody))

		resp, e := client.Execute(ctx, req, httpclient.JsonBody(&EchoResponse{}))
		g.Expect(e).To(Succeed(), "execute request shouldn't fail")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), nil, "", 0)

		expected := EchoResponse{
			Headers: map[string]string{
				"Mockpfx-Ids-Sampled": "",
				"Mockpfx-Ids-Spanid": "",
				"Mockpfx-Ids-Traceid": "",
			},
			ReqBody: reqBody,
		}
		assertResponse(t, g, resp, http.StatusOK, &expected)
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

func ExpectedOpName(method string) string {
	return fmt.Sprintf(`remote-http %s`, method)
}

func AssertSpans(ctx context.Context, g *gomega.WithT, spans []*mocktracer.MockSpan, expectedParent *mocktracer.MockSpan, expectedMethod string, expectedSC int) *mocktracer.MockSpan {
	span := Last(spans)
	g.Expect(span).ToNot(BeNil(), "recorded span should be available")
	if expectedParent != nil {
		g.Expect(span.OperationName).To(Equal(ExpectedOpName(expectedMethod)), "recorded span should have correct '%s'", "OpName")
		g.Expect(span.SpanContext.TraceID).To(Equal(expectedParent.SpanContext.TraceID), "recorded span should have correct '%s'", "TraceID")
		g.Expect(span.ParentID).To(Equal(expectedParent.SpanContext.SpanID), "recorded span should have correct '%s'", "ParentID")
		g.Expect(span.Tags()).To(HaveKeyWithValue("span.kind", BeEquivalentTo("client")), "recorded span should have correct '%s'", "Tags")
		g.Expect(span.Tags()).To(HaveKeyWithValue("method", expectedMethod), "recorded span should have correct '%s'", "Tags")
		g.Expect(span.Tags()).To(HaveKeyWithValue("http.method", expectedMethod), "recorded span should have correct '%s'", "Tags")
		g.Expect(span.Tags()).To(HaveKey("url"), "recorded span should have correct '%s'", "Tags")
		g.Expect(span.Tags()).To(HaveKey("http.url"), "recorded span should have correct '%s'", "Tags")
		g.Expect(span.Tags()).To(HaveKeyWithValue("peer.hostname", "localhost"), "recorded span should have correct '%s'", "Tags")
		g.Expect(span.Tags()).To(HaveKeyWithValue("peer.port", BeEquivalentTo(webtest.CurrentPort(ctx))), "recorded span should have correct '%s'", "Tags")
		g.Expect(span.Tags()).To(HaveKeyWithValue("sc", expectedSC), "recorded span should have correct '%s'", "Tags")
		g.Expect(span.Tags()).To(HaveKeyWithValue("http.status_code", BeEquivalentTo(expectedSC)), "recorded span should have correct '%s'", "Tags")
	} else {
		// because we use real server, server-side tracing is always enabled.
		g.Expect(span.ParentID).To(BeZero(), "recorded span should have correct '%s'", "ParentID")
	}

	return span
}
