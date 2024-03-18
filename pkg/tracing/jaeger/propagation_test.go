package jaegertracing

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/fx"
	"io"
	"net/http"
	"testing"
)

/*************************
	Setup Test
 *************************/

const B3RegExp = `[0-9a-f]+-[0-9a-f]+-1(-[0-9a-f]+)?`
const B3IDRegExp = `[0-9a-f]+`

const (
	KeyB3SingleHeader = `b3`
	KeyB3HttpTraceID = `X-B3-Traceid`
	KeyB3HttpSpanID = `X-B3-Spanid`
	KeyB3HttpParentID = `X-B3-Parentspanid`
	KeyB3HttpSampledID = `X-B3-Sampled`
)

/*************************
	Tests
 *************************/

type TestTracerDI struct {
	fx.In
	AppContext *bootstrap.ApplicationContext
}

func TestSpanOperator(t *testing.T) {
	di := TestTracerDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestInjectTextMap(&di), "TestInjectTextMap"),
		test.GomegaSubTest(SubTestExtractTextMap(&di), "TestExtractTextMap"),
		test.GomegaSubTest(SubTestInjectHTTPHeaders(&di), "TestInjectHTTPHeaders"),
		test.GomegaSubTest(SubTestExtractHttpHeaders(&di), "TestExtractHttpHeaders"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestInjectTextMap(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		ctx = tracing.WithTracer(tracer).ForceNewSpan(ctx)
		span := tracing.SpanFromContext(ctx)
		carrier := map[string]string{}
		e := tracer.Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(carrier))
		g.Expect(e).To(Succeed(), "Inject should not fail")
		g.Expect(carrier).To(HaveKeyWithValue(KeyB3SingleHeader, MatchRegexp(B3RegExp)),
			"Carrier should have correct value")

		ctx = tracing.WithTracer(tracer).DescendantOrNoSpan(ctx)
		span = tracing.SpanFromContext(ctx)
		carrier = map[string]string{}
		e = tracer.Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(carrier))
		g.Expect(e).To(Succeed(), "Inject should not fail")
		g.Expect(carrier).To(HaveKeyWithValue(KeyB3SingleHeader, MatchRegexp(B3RegExp)),
			"Carrier should have correct value")
	}
}

func SubTestExtractTextMap(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		carrier := map[string]string{
			KeyB3SingleHeader: "2d25941072d66722-1d25941072d66711-1",
		}
		extracted, e := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(carrier))
		g.Expect(e).To(Succeed(), "Extract should not fail")
		spanCtx := extracted.(jaeger.SpanContext)
		AssertSpanContextValue(g, spanCtx.TraceID(), jaegerTraceIdString, "2d25941072d66722", "TraceID")
		AssertSpanContextValue(g, spanCtx.SpanID(), jaegerSpanIdString, "1d25941072d66711", "SpanID")
		g.Expect(spanCtx.ParentID()).To(BeZero(), "ParentID should be correct")
		g.Expect(spanCtx.IsSampled()).To(BeTrue(), "IsSampled should be correct")

		carrier = map[string]string{
			KeyB3SingleHeader: "2d25941072d66722-1d25941072d66711-1-2d25941072d66722",
		}
		extracted, e = tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(carrier))
		g.Expect(e).To(Succeed(), "Extract should not fail")
		spanCtx = extracted.(jaeger.SpanContext)
		AssertSpanContextValue(g, spanCtx.TraceID(), jaegerTraceIdString, "2d25941072d66722", "TraceID")
		AssertSpanContextValue(g, spanCtx.SpanID(), jaegerSpanIdString, "1d25941072d66711", "SpanID")
		AssertSpanContextValue(g, spanCtx.ParentID(), jaegerSpanIdString, "2d25941072d66722", "ParentID")
		g.Expect(spanCtx.IsSampled()).To(BeTrue(), "IsSampled should be correct")
	}
}

func SubTestInjectHTTPHeaders(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		ctx = tracing.WithTracer(tracer).ForceNewSpan(ctx)
		span := tracing.SpanFromContext(ctx)
		carrier := http.Header{}
		e := tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(carrier))
		g.Expect(e).To(Succeed(), "Inject should not fail")
		g.Expect(carrier).To(HaveKeyWithValue(KeyB3HttpTraceID, HaveExactElements(MatchRegexp(B3IDRegExp))), "Carrier should have correct value")
		g.Expect(carrier).To(HaveKeyWithValue(KeyB3HttpSpanID, HaveExactElements(MatchRegexp(B3IDRegExp))), "Carrier should have correct value")
		g.Expect(carrier).To(HaveKeyWithValue(KeyB3HttpSampledID, HaveExactElements("1")), "Carrier should have correct value")
		g.Expect(carrier).ToNot(HaveKey(KeyB3HttpParentID), "Carrier should have correct value")

		ctx = tracing.WithTracer(tracer).DescendantOrNoSpan(ctx)
		span = tracing.SpanFromContext(ctx)
		carrier = http.Header{}
		e = tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(carrier))
		g.Expect(e).To(Succeed(), "Inject should not fail")
		g.Expect(carrier).To(HaveKeyWithValue(KeyB3HttpTraceID, HaveExactElements(MatchRegexp(B3IDRegExp))), "Carrier should have correct value")
		g.Expect(carrier).To(HaveKeyWithValue(KeyB3HttpSpanID, HaveExactElements(MatchRegexp(B3IDRegExp))), "Carrier should have correct value")
		g.Expect(carrier).To(HaveKeyWithValue(KeyB3HttpSampledID, HaveExactElements("1")), "Carrier should have correct value")
		g.Expect(carrier).To(HaveKeyWithValue(KeyB3HttpParentID, HaveExactElements(MatchRegexp(B3IDRegExp))), "Carrier should have correct value")
	}
}

func SubTestExtractHttpHeaders(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		carrier := http.Header{
			KeyB3HttpTraceID: []string{"2d25941072d66722"},
			KeyB3HttpSpanID: []string{"1d25941072d66711"},
			KeyB3HttpSampledID: []string{"1"},
			KeyB3HttpParentID: []string{"2d25941072d66722"},
		}
		extracted, e := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(carrier))
		g.Expect(e).To(Succeed(), "Extract should not fail")
		spanCtx := extracted.(jaeger.SpanContext)
		AssertSpanContextValue(g, spanCtx.TraceID(), jaegerTraceIdString, "2d25941072d66722", "TraceID")
		AssertSpanContextValue(g, spanCtx.SpanID(), jaegerSpanIdString, "1d25941072d66711", "SpanID")
		AssertSpanContextValue(g, spanCtx.ParentID(), jaegerSpanIdString, "2d25941072d66722", "ParentID")
		g.Expect(spanCtx.IsSampled()).To(BeTrue(), "IsSampled should be correct")
	}
}

/*************************
	Helper
 *************************/

func NewTestTracer(appCtx *bootstrap.ApplicationContext) (opentracing.Tracer, io.Closer) {
	props := tracing.NewTracingProperties()
	// note: tags is only injected when the span is sampled
	props.Sampler.Enabled = true
	props.Zipkin.Enabled = true
	return NewTracer(appCtx, &props.Jaeger, &props.Sampler)
}

func AssertSpanContextValue[T any](g *gomega.WithT, v T, stringer func(T)string, expected string, name string) {
	g.Expect(stringer(v)).To(Equal(expected), `%s should be correct`, name)
}
