package data_test

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/dbtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"go.uber.org/fx"
	"gorm.io/gorm"
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

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		dbtest.EnableDBRecordMode(),
//	)
//}

type TestTracingDI struct {
	fx.In
	dbtest.DI
	MockTracer *mocktracer.MockTracer
}

func TestGormTracingWithExistingSpan(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithFxOptions(
			fx.Provide(NewTestTracer),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupWithTable(&di.DI)),
		test.SubTestSetup(SetupTestResetTracer(&di)),
		test.SubTestSetup(SetupTestStartSpan(&di)),
		test.SubTestTeardown(TeardownWithTruncateTable(&di.DI)),
		test.GomegaSubTest(SubTestCreate(&di), "TestCreate"),
		test.GomegaSubTest(SubTestRead(&di), "TestRead"),
		test.GomegaSubTest(SubTestUpdate(&di), "TestUpdate"),
		test.GomegaSubTest(SubTestDelete(&di), "TestDelete"),
		test.GomegaSubTest(SubTestRawSQL(&di), "TestRawSQL"),
	)
}

func TestGormTracingWithoutExistingSpan(t *testing.T) {
	di := TestTracingDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		dbtest.WithDBPlayback("testdb"),
		apptest.WithFxOptions(
			fx.Provide(NewTestTracer),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupWithTable(&di.DI)),
		test.SubTestSetup(SetupTestResetTracer(&di)),
		test.SubTestTeardown(TeardownWithTruncateTable(&di.DI)),
		test.GomegaSubTest(SubTestCreate(&di), "TestCreate"),
		test.GomegaSubTest(SubTestRead(&di), "TestRead"),
		test.GomegaSubTest(SubTestUpdate(&di), "TestUpdate"),
		test.GomegaSubTest(SubTestDelete(&di), "TestDelete"),
		test.GomegaSubTest(SubTestRawSQL(&di), "TestRawSQL"),
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

func SubTestCreate(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		var m TestModel
		span := FindSpan(ctx)
		m = TestModel{
			ID:        TestModelID3,
			UniqueKey: "Model-3",
			Value:     "what ever",
		}
		rs = di.DB.WithContext(ctx).Create(&m)
		g.Expect(rs.Error).To(Succeed(), "create model [%s] should not fail", m.UniqueKey)
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "create", false)

		m = TestModel{
			ID:        TestModelID1,
			UniqueKey: "Model-1", // duplicate key
			Value:     "what ever",
		}
		rs = di.DB.WithContext(ctx).Create(&m)
		g.Expect(rs.Error).To(HaveOccurred(), "create model [%s] should fail", m.UniqueKey)
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "create", true)
	}
}

func SubTestRead(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		span := FindSpan(ctx)
		var m []*TestModel
		rs = di.DB.WithContext(ctx).Find(&m)
		g.Expect(rs.Error).To(Succeed(), "read model should not fail")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "select", false)

		rs = di.DB.WithContext(ctx).Find(&m, "non_exist_field = ?", "whatever")
		g.Expect(rs.Error).To(HaveOccurred(), "read model should fail")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "select", true)
	}
}

func SubTestUpdate(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		span := FindSpan(ctx)
		rs = di.DB.WithContext(ctx).Model(&TestModel{ID: TestModelID1}).Updates(&TestModel{Value: "updated"})
		g.Expect(rs.Error).To(Succeed(), "update model should not fail")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "update", false)

		rs = di.DB.WithContext(ctx).Model(&TestModel{ID: TestModelID1}).UpdateColumn("not_exist", "whatever")
		g.Expect(rs.Error).To(HaveOccurred(), "update model should fail")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "update", true)
	}
}

func SubTestDelete(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		span := FindSpan(ctx)
		rs = di.DB.WithContext(ctx).Delete(&TestModel{ID: TestModelID1})
		g.Expect(rs.Error).To(Succeed(), "delete model should not fail")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "delete", false)

		rs = di.DB.WithContext(ctx).Delete(&TestModel{})
		g.Expect(rs.Error).To(HaveOccurred(), "delete model should fail")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "delete", true)
	}
}

func SubTestRawSQL(di *TestTracingDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var rs *gorm.DB
		span := FindSpan(ctx)
		rs = di.DB.WithContext(ctx).Exec(`SELECT id FROM test_model;`)
		g.Expect(rs.Error).To(Succeed(), "raw SQL should not fail")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "sql", false)

		rs = di.DB.WithContext(ctx).Exec(`SELECT unknown_col FROM test_model;`)
		g.Expect(rs.Error).To(HaveOccurred(), "raw SQL should fail")
		AssertSpans(ctx, g, di.MockTracer.FinishedSpans(), span, "sql", true)
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
	return fmt.Sprintf(`db %s`, op)
}

func AssertSpans(_ context.Context, g *gomega.WithT, spans []*mocktracer.MockSpan, expectedParent *mocktracer.MockSpan, expectedOp string, expectErr bool) *mocktracer.MockSpan {
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
	if expectedOp != "sql" {
		g.Expect(span.Tags()).To(HaveKeyWithValue("table", TestModel{}.TableName()), "recorded span should have correct '%s'", "Tags")
	}
	if expectErr {
		g.Expect(span.Tags()).To(HaveKeyWithValue("err", Not(BeZero())), "recorded span should have correct '%s'", "Tags")
	} else {
		g.Expect(span.Tags()).To(HaveKeyWithValue("rows", BeNumerically(">=", 0)), "recorded span should have correct '%s'", "Tags")
	}
	return span
}
