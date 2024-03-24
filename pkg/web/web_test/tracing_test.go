package web_test

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/rest"
	webtracing "github.com/cisco-open/go-lanai/pkg/web/tracing"
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
	"testing"
)

const (
	TestPathTraced = `/traced`
	TestPathExcluded = `/health`
)

/*************************
	Test Setup
 *************************/

func NewTestTracer() (opentracing.Tracer, *mocktracer.MockTracer) {
	tracer := mocktracer.New()
	return tracer, tracer
}

func NewTestController() *SpanCapturingController {
	return &SpanCapturingController{}
}

func ProvideWebController(c *SpanCapturingController) web.Controller {
	return c
}

/*************************
	Tests
 *************************/

type TestTracingDI struct {
	fx.In
	Controller *SpanCapturingController
	MockTracer *mocktracer.MockTracer
}

func TestWebTracingWithExistingSpan(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithModules(webtracing.Module),
		apptest.WithFxOptions(
			fx.Provide(NewTestTracer, NewTestController),
			web.FxControllerProviders(ProvideWebController),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestResetTracer(&di)),
		test.GomegaSubTest(SubTestRequestWithPropagationHeaders(&di), "TestRequestWithPropagationHeaders"),
		test.GomegaSubTest(SubTestRequestWithoutHeaders(&di), "SubTestRequestWithoutHeaders"),
		test.GomegaSubTest(SubTestExcludedRequest(&di), "TestExcludedRequest"),
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

func SubTestRequestWithPropagationHeaders(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx, span := ContextWithTestSpan(ctx, di.MockTracer)
		// prepare controller
		var ctxSpan, reqSpan *mocktracer.MockSpan
		di.Controller.CaptureFunc = func(ctx context.Context, req *http.Request) {
			ctxSpan = FindSpan(ctx)
			reqSpan = FindSpan(req.Context())
		}

		// send request
		req := webtest.NewRequest(ctx, http.MethodGet, TestPathTraced, nil, func(req *http.Request) {
			_ = di.MockTracer.Inject(span.SpanContext, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		})
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp.Response).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response's status code should be correct")

		finishedSpan := Last(di.MockTracer.FinishedSpans())
		AssertSpans(ctx, g, finishedSpan, span, TestPathTraced, http.MethodGet, http.StatusOK)
		AssertSpans(ctx, g, ctxSpan, span, TestPathTraced, http.MethodGet, http.StatusOK)
		AssertSpans(ctx, g, reqSpan, span, TestPathTraced, http.MethodGet, http.StatusOK)
	}
}

func SubTestRequestWithoutHeaders(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// prepare controller
		var ctxSpan, reqSpan *mocktracer.MockSpan
		di.Controller.CaptureFunc = func(ctx context.Context, req *http.Request) {
			ctxSpan = FindSpan(ctx)
			reqSpan = FindSpan(req.Context())
		}

		// send request
		req := webtest.NewRequest(ctx, http.MethodGet, TestPathTraced, nil)
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp.Response).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response's status code should be correct")

		finishedSpan := Last(di.MockTracer.FinishedSpans())
		AssertSpans(ctx, g, finishedSpan, nil, TestPathTraced, http.MethodGet, http.StatusOK)
		AssertSpans(ctx, g, ctxSpan, nil, TestPathTraced, http.MethodGet, http.StatusOK)
		AssertSpans(ctx, g, reqSpan, nil, TestPathTraced, http.MethodGet, http.StatusOK)
	}
}

func SubTestExcludedRequest(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// prepare controller
		var ctxSpan, reqSpan *mocktracer.MockSpan
		di.Controller.CaptureFunc = func(ctx context.Context, req *http.Request) {
			ctxSpan = FindSpan(ctx)
			reqSpan = FindSpan(req.Context())
		}

		// send health request
		req := webtest.NewRequest(ctx, http.MethodGet, TestPathExcluded, nil)
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp.Response).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response's status code should be correct")

		finishedSpan := Last(di.MockTracer.FinishedSpans())
		g.Expect(finishedSpan).To(BeNil(), "span should not be created")
		g.Expect(ctxSpan).To(BeNil(), "span should not be created")
		g.Expect(reqSpan).To(BeNil(), "span should not be created")

		// send CORS preflight
		req = webtest.NewRequest(ctx, http.MethodOptions, TestPathTraced, nil)
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp.Response).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response's status code should be correct")

		finishedSpan = Last(di.MockTracer.FinishedSpans())
		g.Expect(finishedSpan).To(BeNil(), "span should not be created")
		g.Expect(ctxSpan).To(BeNil(), "span should not be created")
		g.Expect(reqSpan).To(BeNil(), "span should not be created")
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

func ExpectedOpName(path string) string {
	return fmt.Sprintf(`http /test%s`, path)
}

func AssertSpans(_ context.Context, g *gomega.WithT, span *mocktracer.MockSpan, expectedParent *mocktracer.MockSpan, expectedPath string, expectedMethod string, expectedSC int) *mocktracer.MockSpan {
	g.Expect(span).ToNot(BeNil(), "recorded span should be available")
	g.Expect(span.OperationName).To(Equal(ExpectedOpName(expectedPath)), "recorded span should have correct '%s'", "OpName")
	if expectedParent != nil {
		g.Expect(span.SpanContext.TraceID).To(Equal(expectedParent.SpanContext.TraceID), "recorded span should have correct '%s'", "TraceID")
		g.Expect(span.ParentID).To(Equal(expectedParent.SpanContext.SpanID), "recorded span should have correct '%s'", "ParentID")
	} else {
		g.Expect(span.SpanContext.TraceID).ToNot(BeZero(), "recorded span should have correct '%s'", "TraceID")
		g.Expect(span.ParentID).To(BeZero(), "recorded span should have correct '%s'", "ParentID")
	}
	g.Expect(span.Tags()).To(HaveKeyWithValue("span.kind", BeEquivalentTo("server")), "recorded span should have correct '%s'", "Tags")
	g.Expect(span.Tags()).To(HaveKeyWithValue("http.method", expectedMethod), "recorded span should have correct '%s'", "Tags")
	g.Expect(span.Tags()).To(HaveKeyWithValue("http.url", HaveSuffix(expectedPath)), "recorded span should have correct '%s'", "Tags")
	g.Expect(span.Tags()).To(HaveKeyWithValue("http.status_code", BeEquivalentTo(expectedSC)), "recorded span should have correct '%s'", "Tags")
	return span
}

type SpanCapturingController struct {
	CaptureFunc func(ctx context.Context, req *http.Request)
}

func (c *SpanCapturingController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Get("/traced").EndpointFunc(c.Capture).Build(),
		rest.Get("/health").EndpointFunc(c.Capture).Build(),
		rest.Options("/traced").EndpointFunc(c.Capture).Build(),
	}
}

func (c *SpanCapturingController) Capture(ctx context.Context, r *http.Request) (interface{}, error) {
	if c.CaptureFunc != nil {
		c.CaptureFunc(ctx, r)
	}
	return map[string]interface{}{"hello": "stranger"}, nil
}
