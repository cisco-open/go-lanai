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

package tracing_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	jaegertracing "github.com/cisco-open/go-lanai/pkg/tracing/jaeger"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/fx"
	"io"
	"testing"
)

/*************************
	Setup Test
 *************************/

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
		//apptest.WithTimeout(120*time.Second),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestForceNewSpan(&di), "TestNewSpan"),
		test.GomegaSubTest(SubTestNewOrDescendant(&di), "TestNewOrDescendant"),
		test.GomegaSubTest(SubTestNewOrFollows(&di), "TestNewOrFollows"),
		test.GomegaSubTest(SubTestDescendantOrNoop(&di), "TestDescendantOrNoop"),
		test.GomegaSubTest(SubTestFollowsOrNoop(&di), "TestFollowsOrNoop"),
		test.GomegaSubTest(SubTestRewind(&di), "TestRewind"),
		test.GomegaSubTest(SubTestSpanCustomization(&di), "TestSpanCustomization"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestForceNewSpan(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		AssertNoSpan(ctx, g)
		ctx = tracing.WithTracer(tracer).ForceNewSpan(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectRoot()))
	}
}

func SubTestNewOrDescendant(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		AssertNoSpan(ctx, g)
		ctx = tracing.WithTracer(tracer).NewSpanOrDescendant(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectRoot()))

		traceID := tracing.TraceIdFromContext(ctx)
		ctx = tracing.WithTracer(tracer).NewSpanOrDescendant(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectTraceID(traceID), ExpectParentID(traceID)))
	}
}

func SubTestNewOrFollows(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		AssertNoSpan(ctx, g)
		ctx = tracing.WithTracer(tracer).NewSpanOrFollows(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectRoot()))

		traceID := tracing.TraceIdFromContext(ctx)
		ctx = tracing.WithTracer(tracer).NewSpanOrFollows(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectTraceID(traceID), ExpectParentID(traceID), ExpectFollowed()))
	}
}

func SubTestDescendantOrNoop(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		AssertNoSpan(ctx, g)
		ctx = tracing.WithTracer(tracer).DescendantOrNoSpan(ctx)
		AssertNoSpan(ctx, g)

		ctx = tracing.WithTracer(tracer).ForceNewSpan(ctx)
		traceID := tracing.TraceIdFromContext(ctx)
		ctx = tracing.WithTracer(tracer).DescendantOrNoSpan(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectTraceID(traceID), ExpectParentID(traceID)))
	}
}

func SubTestFollowsOrNoop(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		AssertNoSpan(ctx, g)
		ctx = tracing.WithTracer(tracer).FollowsOrNoSpan(ctx)
		AssertNoSpan(ctx, g)

		ctx = tracing.WithTracer(tracer).ForceNewSpan(ctx)
		traceID := tracing.TraceIdFromContext(ctx)
		ctx = tracing.WithTracer(tracer).FollowsOrNoSpan(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectTraceID(traceID), ExpectParentID(traceID), ExpectFollowed()))
	}
}

func SubTestRewind(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		AssertNoSpan(ctx, g)
		ctx = tracing.WithTracer(tracer).ForceNewSpan(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectRoot()))
		traceID := tracing.TraceIdFromContext(ctx)

		ctx = tracing.WithTracer(tracer).NewSpanOrDescendant(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectTraceID(traceID), ExpectParentID(traceID)))

		ctx = tracing.WithTracer(tracer).FinishAndRewind(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectTraceID(traceID), ExpectRoot()))

		ctx = tracing.WithTracer(tracer).FinishAndRewind(ctx)
		AssertNoSpan(ctx, g)

		ctx = tracing.WithTracer(tracer).FinishAndRewind(ctx)
		AssertNoSpan(ctx, g)
	}
}

func SubTestSpanCustomization(di *TestTracerDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tracer, closer := NewTestTracer(di.AppContext)
		defer func() { _ = closer.Close() }()
		AssertNoSpan(ctx, g)
		ctx = tracing.WithTracer(tracer).ForceNewSpan(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(ExpectRoot()))

		traceID := tracing.TraceIdFromContext(ctx)
		ctx = tracing.WithTracer(tracer).
			WithOpName("test").
			WithStartOptions(TestTagUpdater{K: "start-tag", V: "test-start-opt"}).
			WithOptions(
				tracing.SpanTag("tag", "test-tag"),
				tracing.SpanKind(ext.SpanKindRPCClientEnum),
				tracing.SpanComponent("test"),
				tracing.SpanBaggageItem("key", "test-baggage"),
				tracing.SpanHttpUrl("http://localhost:0/test"),
				tracing.SpanHttpMethod("GET"),
				tracing.SpanHttpStatusCode(200),
			).
			NewSpanOrDescendant(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(
			ExpectTraceID(traceID),
			ExpectOpName("test"),
			ExpectBaggage("key", "test-baggage"),
			ExpectTag("tag", "test-tag"), ExpectTag("start-tag", "test-start-opt"),
			ExpectTag(string(ext.SpanKind), "client"), ExpectTag(string(ext.Component), "test"),
			ExpectTag(string(ext.HTTPUrl), "http://localhost:0/test"),
			ExpectTag(string(ext.HTTPMethod), "GET"), ExpectTag(string(ext.HTTPStatusCode), 200),
		))

		// try update
		tracing.WithTracer(tracer).
			WithOptions(
			tracing.SpanTag("tag", "updated"),
				tracing.SpanBaggageItem("key", "updated"),
			).
			UpdateCurrentSpan(ctx)
		AssertCurrentSpan(ctx, g, SpanExpectation(
			ExpectTraceID(traceID),
			ExpectOpName("test"),
			ExpectBaggage("key", "updated"),
			ExpectTag("tag", "updated"), ExpectTag("start-tag", "test-start-opt"),
			ExpectTag(string(ext.SpanKind), "client"), ExpectTag(string(ext.Component), "test"),
			ExpectTag(string(ext.HTTPUrl), "http://localhost:0/test"),
			ExpectTag(string(ext.HTTPMethod), "GET"), ExpectTag(string(ext.HTTPStatusCode), 200),
		))
	}
}

/*************************
	Helper
 *************************/

func NewTestTracer(appCtx *bootstrap.ApplicationContext) (opentracing.Tracer, io.Closer) {
	props := tracing.NewTracingProperties()
	// note: tags is only injected when the span is sampled
	props.Sampler.Enabled = true
	return jaegertracing.NewTracer(appCtx, &props.Jaeger, &props.Sampler)
}

func AssertNoSpan(ctx context.Context, g *gomega.WithT) {
	span := tracing.SpanFromContext(ctx)
	g.Expect(span).To(BeNil(), "current span should be nil")
}

type ExpectedSpan struct {
	TraceID    interface{}
	ParentID   interface{}
	IsRoot     bool
	IsFollowed bool
	OpName     string
	Tags       map[string]interface{}
	Baggage    map[string]interface{}
}

func SpanExpectation(opts ...func(expect *ExpectedSpan)) ExpectedSpan {
	expect := ExpectedSpan{
		Tags:    make(map[string]interface{}),
		Baggage: make(map[string]interface{}),
	}
	for _, fn := range opts {
		fn(&expect)
	}
	return expect
}

func ExpectTraceID(v interface{}) func(expect *ExpectedSpan) {
	return func(expect *ExpectedSpan) {
		expect.TraceID = v
	}
}

func ExpectParentID(v interface{}) func(expect *ExpectedSpan) {
	return func(expect *ExpectedSpan) {
		expect.ParentID = v
	}
}

func ExpectRoot() func(expect *ExpectedSpan) {
	return func(expect *ExpectedSpan) {
		expect.IsRoot = true
	}
}

func ExpectFollowed() func(expect *ExpectedSpan) {
	return func(expect *ExpectedSpan) {
		expect.IsFollowed = true
	}
}

func ExpectOpName(name string) func(expect *ExpectedSpan) {
	return func(expect *ExpectedSpan) {
		expect.OpName = name
	}
}

func ExpectBaggage(k string, v interface{}) func(expect *ExpectedSpan) {
	return func(expect *ExpectedSpan) {
		expect.Baggage[k] = v
	}
}

func ExpectTag(k string, v interface{}) func(expect *ExpectedSpan) {
	return func(expect *ExpectedSpan) {
		expect.Tags[k] = v
	}
}

func AssertCurrentSpan(ctx context.Context, g *gomega.WithT, expect ExpectedSpan) {
	span := tracing.SpanFromContext(ctx)
	g.Expect(span).ToNot(BeNil(), "current span should not be nil")
	g.Expect(tracing.SpanIdFromContext(ctx)).ToNot(BeZero(), "span ID should not be zero")
	if expect.TraceID != nil {
		g.Expect(tracing.TraceIdFromContext(ctx)).To(Equal(expect.TraceID), "trace ID should be correct")
	} else {
		g.Expect(tracing.TraceIdFromContext(ctx)).ToNot(BeZero(), "trace ID should be correct")
	}
	switch {
	case expect.IsRoot:
		g.Expect(tracing.ParentIdFromContext(ctx)).To(BeEmpty(), "parent ID should be empty")
	case expect.ParentID != nil:
		g.Expect(tracing.ParentIdFromContext(ctx)).To(Equal(expect.ParentID), "parent ID should be correct")
	default:
		g.Expect(tracing.ParentIdFromContext(ctx)).ToNot(BeEmpty(), "parent ID should not be empty")
	}

	if len(expect.Baggage) != 0 {
		for k, v := range expect.Baggage {
			g.Expect(span.BaggageItem(k)).To(BeEquivalentTo(v), "span should have correct baggage")
		}
	}

	jSpan, ok := span.(*jaeger.Span)
	if !ok {
		return
	}
	switch {
	case expect.IsRoot:
		g.Expect(jSpan.References()).To(HaveLen(0), "references should be empty")
	case expect.IsFollowed:
		g.Expect(jSpan.References()).ToNot(BeEmpty(), "references should not be empty")
		refs := jSpan.References()
		ref := refs[len(refs)-1]
		g.Expect(ref.Type).To(Equal(opentracing.FollowsFromRef))
	default:
		g.Expect(jSpan.References()).ToNot(BeEmpty(), "references should not be empty")
		refs := jSpan.References()
		ref := refs[len(refs)-1]
		g.Expect(ref.Type).To(Equal(opentracing.ChildOfRef))
	}

	if len(expect.OpName) != 0 {
		g.Expect(jSpan.OperationName()).To(Equal(expect.OpName), "op name should be correct")
	}

	if len(expect.Tags) != 0 {
		tags := jSpan.Tags()
		for k, v := range expect.Tags {
			g.Expect(tags).To(HaveKeyWithValue(k, BeEquivalentTo(v)), "span tags should contain %s=%v", k, v)
		}
	}
}

type TestTagUpdater struct {
	K, V string
}

func (u TestTagUpdater) Apply(opts *opentracing.StartSpanOptions) {
	if opts.Tags == nil {
		opts.Tags = make(map[string]interface{})
	}
	opts.Tags[u.K] = u.V
}
