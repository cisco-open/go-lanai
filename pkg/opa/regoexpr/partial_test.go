package regoexpr

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"fmt"
	"github.com/onsi/gomega"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"strings"
	"testing"
)

/*************************
	Test Setup
 *************************/

const ModuleTemplate = `
package test
import future.keywords
%s
`
const (
	ModulePackage  = "test"
	ModuleFileName = "test.rego"
	TargetValue    = "target"
)

var DefaultUnknowns = []string{"input.list", "input.map", "input.map_list", "input.single"}

/*************************
	Test
 *************************/

func TestPartialParser(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestEqual(), "TestEqual"),
		test.GomegaSubTest(SubTestNotEqual(), "TestNotEqual"),
		test.GomegaSubTest(SubTestNegate(), "TestNegate"),
		test.GomegaSubTest(SubTestComparison(), "TestComparison"),
		test.GomegaSubTest(SubTestContains(), "TestContains"),
		test.GomegaSubTest(SubTestAnd(), "TestAnd"),
		test.GomegaSubTest(SubTestMultiQueries(), "TestMultiQueries"),
	)
}

func TestNormalizeQueries(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestDuplicateQueries(), "TestDuplicateQueries"),
		test.GomegaSubTest(SubTestDuplicateExprs(), "TestDuplicateExprs"),
		test.GomegaSubTest(SubTestMultiValuesOnUnknown(), "TestMultiValuesOnUnknown"),
		test.GomegaSubTest(SubTestEqualAndNotEqual(), "TestEqualAndNotEqual"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestEqual() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if input.single = "target"`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, Expect("input.single", "eq", TargetValue))
	}
}

func SubTestNotEqual() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if input.single != "target"`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, Expect("input.single", "neq", TargetValue))
	}
}

func SubTestNegate() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if not input.single = "target"`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, Expect("input.single", "neq", TargetValue))
	}
}

func SubTestComparison() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if input.single > 0`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, Expect("input.single", "gt", 0))
	}
}

func SubTestContains() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if input.list[_] = "target"`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, Expect("input.list", "internal.in", TargetValue))
	}
}

func SubTestAnd() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			input.list[_] = "target"
			input.map["map-key"] = "target"
			input.map_list["map-key"][_] = "target"
		}`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, ExpectAnd(
			Expect(`input.list`, "internal.in", TargetValue),
			Expect(`input.map["map-key"]`, "eq", TargetValue),
			Expect(`input.map_list["map-key"]`, "internal.in", TargetValue),
		))
	}
}

func SubTestMultiQueries() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			input.list[_] = "target"
		}
		allow if {
			input.map["map-key"] = "target"
		}
		allow if {
			input.list[_] = "target"
			input.map["map-key"] = "target"
			input.map_list["map-key"][_] = "target"
		}`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs,
			Expect(`input.list`, "internal.in", TargetValue),
			Expect(`input.map["map-key"]`, "eq", TargetValue),
			ExpectAnd(
				Expect(`input.list`, "internal.in", TargetValue),
				Expect(`input.map["map-key"]`, "eq", TargetValue),
				Expect(`input.map_list["map-key"]`, "internal.in", TargetValue),
			),
		)
	}
}

func SubTestDuplicateQueries() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			input.single = "target"
		}
		allow if {
			input.single = "target"
		}
		allow if {
			input.list[_] = "target"
			input.map["map-key"] = "target"
			input.map_list["map-key"][_] = "target"
		}`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		// Note: for some reason, the query with duplicates always appears at the end.
		AssertExpressions(ctx, g, exprs,
			ExpectAnd(
				Expect(`input.list`, "internal.in", TargetValue),
				Expect(`input.map["map-key"]`, "eq", TargetValue),
				Expect(`input.map_list["map-key"]`, "internal.in", TargetValue),
			),
			Expect(`input.single`, "eq", TargetValue),
		)
	}
}

func SubTestDuplicateExprs() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			input.single = "target"
			input.list[_] = "target"
			input.single = "target"
		}
		allow if {
			input.list[_] = "target"
		}`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		// Note: for some reason, the query with duplicates always appears at the end.
		AssertExpressions(ctx, g, exprs,
			Expect(`input.list`, "internal.in", TargetValue),
			ExpectAnd(
				Expect(`input.single`, "eq", TargetValue),
				Expect(`input.list`, "internal.in", TargetValue),
			),
		)
	}
}

func SubTestMultiValuesOnUnknown() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			input.single = "target"
			input.list[_] = "target"
			input.single = "another value"
		}
		allow if {
			input.list[_] = "target"
			input.list[_] = "another"
		}`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs,
			ExpectAnd(
				Expect(`input.list`, "internal.in", TargetValue),
				Expect(`input.list`, "internal.in", "another"),
			),
		)
	}
}

func SubTestEqualAndNotEqual() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			input.single = "target"
			input.list[_] = "target"
			input.single != "another"
			input.single != "target"
		}
		allow if {
			input.single = "target"
			input.list[_] = "target"
			input.single != "another"
		}`
		exprs := TranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs,
			ExpectAnd(
				Expect(`input.single`, "eq", TargetValue),
				Expect(`input.list`, "internal.in", TargetValue),
				Expect(`input.single`, "neq", "another"),
			),
		)
	}
}

/*************************
	Helpers
 *************************/

type ExpectedExpr struct {
	Exprs []*ExpectedExpr
	Ref   string
	Value interface{}
	Op    string
}

func Expect(ref, op string, value interface{}) *ExpectedExpr {
	return &ExpectedExpr{
		Ref:   ref,
		Value: value,
		Op:    op,
	}
}

func ExpectAnd(exprs ...*ExpectedExpr) *ExpectedExpr {
	return &ExpectedExpr{
		Exprs: exprs,
		Op:    OpAnd,
	}
}

func AssertExpressions(ctx context.Context, g *gomega.WithT, exprs []TestExpression, expected ...*ExpectedExpr) {
	g.Expect(exprs).To(gomega.HaveLen(len(expected)), "expressions should have correct count")
	for i := range exprs {
		expect := expected[i]
		if expect.Op != OpAnd {
			expect = ExpectAnd(expect)
		}
		exprs[i].Assert(ctx, g, expect)
	}
}

func TranslatePartial(ctx context.Context, g *gomega.WithT, rules, queryName string, unknowns ...string) []TestExpression {
	pq := ExecutePartial(ctx, g, rules, queryName, unknowns...)
	exprs, e := TranslatePartialQueries(ctx, pq, func(opts *TranslateOption[TestExpression]) {
		opts.Translator = TestQueryTranslator{}
	})
	g.Expect(e).To(gomega.Succeed(), "translating partial queries should not return error")
	fmt.Printf("Exprs: %v\n", &TestCompositeExpr{
		Exprs: exprs,
		Op:    OpOr,
	})
	return exprs
}

func ExecutePartial(ctx context.Context, g *gomega.WithT, rules, queryName string, unknowns ...string) *rego.PartialQueries {
	module := fmt.Sprintf(ModuleTemplate, rules)
	query := fmt.Sprintf("data.%s.%s", ModulePackage, queryName)
	if len(unknowns) == 0 {
		unknowns = DefaultUnknowns
	}

	prepared, e := rego.New(
		rego.Query(query),
		rego.Unknowns(unknowns),
		rego.Module(ModuleFileName, module),
	).PrepareForPartial(ctx)
	g.Expect(e).To(gomega.Succeed(), "preparing partial should not return error")

	pq, e := prepared.Partial(ctx)
	g.Expect(e).To(gomega.Succeed(), "partial evaluation should not return error")
	g.Expect(pq).ToNot(gomega.BeNil(), "partial queries should not be nil")
	return pq
}

/****************************
	Test Query Translator
 ****************************/

const (
	OpOr    = `OR`
	OpAnd   = `AND`
	OpNotIn = `internal.not_in`
)

type TestExpression interface {
	Assert(ctx context.Context, g *gomega.WithT, expected *ExpectedExpr)
}

type TestExpr struct {
	Ref   string
	Value interface{}
	Op    string
}

func (expr *TestExpr) Assert(_ context.Context, g *gomega.WithT, expected *ExpectedExpr) {
	g.Expect(expected.Exprs).To(gomega.HaveLen(0), "expect %s, but got comparison expression [%v]", expected.Op, expr)
	g.Expect(expr.Op).To(gomega.BeEquivalentTo(expected.Op), "op should be correct")
	g.Expect(expr.Ref).To(gomega.BeEquivalentTo(expected.Ref), "ref should be correct")
	g.Expect(expr.Value).To(gomega.BeEquivalentTo(expected.Value), "value should be correct")
}

func (expr *TestExpr) String() string {
	return fmt.Sprintf("%s %s '%v'", expr.Ref, expr.Op, expr.Value)
}

type TestCompositeExpr struct {
	Exprs []TestExpression
	Op    string
}

func (expr *TestCompositeExpr) Assert(ctx context.Context, g *gomega.WithT, expected *ExpectedExpr) {
	g.Expect(expr.Exprs).To(gomega.HaveLen(len(expected.Exprs)), "composite expressions should have correct length")
	g.Expect(expr.Op).To(gomega.BeEquivalentTo(expected.Op), "op should be correct")
	for i := range expr.Exprs {
		expr.Exprs[i].Assert(ctx, g, expected.Exprs[i])
	}
}

func (expr *TestCompositeExpr) String() (ret string) {
	strs := make([]string, len(expr.Exprs))
	for i := range expr.Exprs {
		strs[i] = fmt.Sprint(expr.Exprs[i])
	}
	switch expr.Op {
	case OpOr:
		ret = "(" + strings.Join(strs, " "+expr.Op+" ") + ")"
	case OpAnd:
		ret = strings.Join(strs, " "+expr.Op+" ")
	}
	return
}

type TestQueryTranslator struct{}

func (t TestQueryTranslator) Negate(ctx context.Context, expr TestExpression) TestExpression {
	switch v := expr.(type) {
	case *TestExpr:
		return t.negate(v)
	case *TestCompositeExpr:
		exprs := make([]TestExpression, len(v.Exprs))
		for i := range v.Exprs {
			exprs[i] = t.Negate(ctx, v.Exprs[i])
		}
		op := OpAnd
		if v.Op == OpAnd {
			op = OpOr
		}
		return &TestCompositeExpr{
			Exprs: exprs,
			Op:    op,
		}
	}
	return expr
}

func (t TestQueryTranslator) And(_ context.Context, exprs ...TestExpression) TestExpression {
	return &TestCompositeExpr{
		Exprs: exprs,
		Op:    OpAnd,
	}
}

func (t TestQueryTranslator) Or(_ context.Context, exprs ...TestExpression) TestExpression {
	return &TestCompositeExpr{
		Exprs: exprs,
		Op:    OpOr,
	}
}

func (t TestQueryTranslator) Comparison(_ context.Context, op ast.Ref, colRef ast.Ref, val interface{}) (ret TestExpression, err error) {
	return &TestExpr{
		Ref:   colRef.String(),
		Value: val,
		Op:    op.String(),
	}, nil
}

func (t TestQueryTranslator) negate(expr *TestExpr) TestExpression {
	newExpr := *expr
	switch expr.Op {
	case OpEq.String(), OpEqual.String():
		newExpr.Op = OpNeq.String()
	case OpNeq.String():
		newExpr.Op = OpEq.String()
	case OpLte.String():
		newExpr.Op = OpGt.String()
	case OpLt.String():
		newExpr.Op = OpGte.String()
	case OpGte.String():
		newExpr.Op = OpLt.String()
	case OpGt.String():
		newExpr.Op = OpLte.String()
	case OpIn.String():
		newExpr.Op = OpNotIn
	default:
		newExpr.Op = "unsupported negation"
	}
	return &newExpr
}
