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

package regoexpr

import (
    "context"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/opa"
    "github.com/cisco-open/go-lanai/test"
    "github.com/onsi/gomega"
    "github.com/open-policy-agent/opa/rego"
    "sort"
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
var DefaultInput = map[string]interface{}{
	"known_list":     []string{"value1", "value2"},
	"known_map":      map[string]string{"key1": "value1", "key2": "value2"},
	"known_map_list": map[string][]string{"key1": {"value1"}, "key2": {"value2"}},
	"known_single":   "value",
}

/*************************
	Test
 *************************/

func TestParseQueries(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestEqual(), "TestEqual"),
		test.GomegaSubTest(SubTestNotEqual(), "TestNotEqual"),
		test.GomegaSubTest(SubTestNegate(), "TestNegate"),
		test.GomegaSubTest(SubTestComparison(), "TestComparison"),
		test.GomegaSubTest(SubTestContains(), "TestContains"),
		test.GomegaSubTest(SubTestAnd(), "TestAnd"),
		test.GomegaSubTest(SubTestMultiQueries(), "TestMultiQueries"),
		test.GomegaSubTest(SubTestAccessDenied(), "TestAccessDenied"),
		test.GomegaSubTest(SubTestAccessGranted(), "TestAccessGranted"),
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

func TestErrorHandling(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestInternalFunctions(), "TestInternalFunctions"),
		test.GomegaSubTest(SubTestNonOperationExpr(), "TestSingleTermOperation"),
		test.GomegaSubTest(SubTestNonThreeTermOperation(), "TestNonThreeTermOperation"),
		test.GomegaSubTest(SubTestUnresolvableValues(), "TestUnresolvableValue"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestEqual() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if input.single = "target"`
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, Expect("input.single", "eq", TargetValue))
	}
}

func SubTestNotEqual() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if input.single != "target"`
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, Expect("input.single", "neq", TargetValue))
	}
}

func SubTestNegate() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if not input.single = "target"`
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, Expect("input.single", "neq", TargetValue))
	}
}

func SubTestComparison() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if input.single > 0`
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs, Expect("input.single", "gt", 0))
	}
}

func SubTestContains() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if input.list[_] = "target"`
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
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
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
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
			input.known_list[_] = "value1"
		}
		allow if {
			input.single = "target"
			input.known_single = "won't match"
		}
		allow if {
			input.map["map-key"] = "target"
			input.known_single = "value"
		}
		allow if {
			input.list[_] = "target"
			input.map["map-key"] = "target"
			input.map_list["map-key"][_] = "target"
		}`
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
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

func SubTestAccessDenied() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			input.list[_] = "target"
			input.known_single = "won't match"
		}
		allow if {
			input.single = "target"
			input.known_list[_] = "won't match"
		}
		allow if {
			input.list[_] = "target"
			input.map["map-key"] = "target"
			input.known_map_list["map-key"][_] = "won't match'"
		}`
		_, e := TranslatePartial(ctx, g, rules, "allow")
		g.Expect(e).To(gomega.HaveOccurred(), "translating partial queries should return error")
		g.Expect(errors.Is(e, opa.ErrQueriesNotResolved)).To(gomega.BeTrue(), "translating partial queries should return ErrQueriesNotResolved")
	}
}

func SubTestAccessGranted() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			input.list[_] = "target"
			input.known_single = "value"
		}
		allow if {
			input.known_list[_] = "value1"
		}
		allow if {
			input.list[_] = "target"
			input.map["map-key"] = "target"
			input.known_map_list["key1"][_] = "value1"
		}`
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
		g.Expect(exprs).ToNot(gomega.BeNil(), "translated expressions should not be nil")
		g.Expect(exprs).To(gomega.HaveLen(0), "granted access should be a empty expressions list")
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
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs,
			Expect(`input.single`, "eq", TargetValue),
			ExpectAnd(
				Expect(`input.list`, "internal.in", TargetValue),
				Expect(`input.map["map-key"]`, "eq", TargetValue),
				Expect(`input.map_list["map-key"]`, "internal.in", TargetValue),
			),
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
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs,
			Expect(`input.list`, "internal.in", TargetValue),
			ExpectAnd(
				Expect(`input.list`, "internal.in", TargetValue),
				Expect(`input.single`, "eq", TargetValue),
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
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs,
			ExpectAnd(
				Expect(`input.list`, "internal.in", "another"),
				Expect(`input.list`, "internal.in", TargetValue),
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
		exprs := MustTranslatePartial(ctx, g, rules, "allow")
		AssertExpressions(ctx, g, exprs,
			ExpectAnd(
				Expect(`input.list`, "internal.in", TargetValue),
				Expect(`input.single`, "eq", TargetValue),
				Expect(`input.single`, "neq", "another"),
			),
		)
	}
}

func SubTestInternalFunctions() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			"target" in input.list
		}
		allow if {
			input.single = "target"
		}`
		_, e := TranslatePartial(ctx, g, rules, "allow")
		g.Expect(e).To(gomega.HaveOccurred(), "translating partial queries should return error")
		g.Expect(errors.Is(e, ParsingError)).To(gomega.BeTrue(), "translating partial queries should return ParsingError")
	}
}

func SubTestNonOperationExpr() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			with_default
		}
		default with_default = false
		with_default if input.single = "target"`
		_, e := TranslatePartial(ctx, g, rules, "allow")
		g.Expect(e).To(gomega.HaveOccurred(), "translating partial queries should return error")
		g.Expect(errors.Is(e, ParsingError)).To(gomega.BeTrue(), "translating partial queries should return ParsingError")
	}
}

func SubTestNonThreeTermOperation() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const rules = `allow if {
			substring(input.single, 2, 3) = "get"
		}`
		_, e := TranslatePartial(ctx, g, rules, "allow")
		g.Expect(e).To(gomega.HaveOccurred(), "translating partial queries should return error")
		g.Expect(errors.Is(e, ParsingError)).To(gomega.BeTrue(), "translating partial queries should return ParsingError")
	}
}

func SubTestUnresolvableValues() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		const rules1 = `allow {
			input.map = call(input.single)
		}
		call(v) := result { result := {"foo": "bar"} }`
		_, e = TranslatePartial(ctx, g, rules1, "allow")
		g.Expect(e).To(gomega.HaveOccurred(), "translating partial queries should return error")
		g.Expect(errors.Is(e, ParsingError)).To(gomega.BeTrue(), "translating partial queries should return ParsingError")

		const rules2 = `allow {
			input.map = call(input.single)
		}
		call(v) := result { result := {"foo": v} }`
		_, e = TranslatePartial(ctx, g, rules2, "allow")
		g.Expect(e).To(gomega.HaveOccurred(), "translating partial queries should return error")
		g.Expect(errors.Is(e, ParsingError)).To(gomega.BeTrue(), "translating partial queries should return ParsingError")

		const rules3 = `allow {
			input.map = with_default
		}
		default with_default = {}
		with_default := {"foo":"bar"} { input.single = "target" }`
		_, e = TranslatePartial(ctx, g, rules3, "allow")
		g.Expect(e).To(gomega.HaveOccurred(), "translating partial queries should return error")
		g.Expect(errors.Is(e, ParsingError)).To(gomega.BeTrue(), "translating partial queries should return ParsingError")
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

func (ee *ExpectedExpr) Compare(another *ExpectedExpr) int {
	switch ee.Op {
	case OpAnd, OpOr:
		if ret := len(ee.Exprs) - len(another.Exprs); ret != 0 {
			return ret
		}
	default:
		if ret := strings.Compare(ee.Ref, another.Ref); ret != 0 {
			return ret
		}
	}
	if ret := strings.Compare(ee.Op, another.Op); ret != 0 {
		return ret
	}
	return 0
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
	for i := range expected {
		if expected[i].Op != OpAnd {
			expected[i] = ExpectAnd(expected[i])
		}
	}
	sort.SliceStable(expected, func(i, j int) bool {
		return expected[i].Compare(expected[j]) < 0
	})
	for i := range exprs {
		exprs[i].Assert(ctx, g, expected[i])
	}
}

func MustTranslatePartial(ctx context.Context, g *gomega.WithT, rules, queryName string, unknowns ...string) []TestExpression {
	exprs, e := TranslatePartial(ctx, g, rules, queryName, unknowns...)
	g.Expect(e).To(gomega.Succeed(), "translating partial queries should not return error")
	return exprs
}

func TranslatePartial(ctx context.Context, g *gomega.WithT, rules, queryName string, unknowns ...string) ([]TestExpression, error) {
	pq := ExecutePartial(ctx, g, rules, queryName, unknowns...)
	exprs, e := TranslatePartialQueries(ctx, pq, func(opts *TranslateOption[TestExpression]) {
		opts.Translator = TestQueryTranslator{}
	})
	if e != nil {
		return nil, e
	}

	sort.SliceStable(exprs, func(i, j int) bool {
		return exprs[i].Compare(exprs[j]) < 0
	})
	fmt.Printf("Exprs: %v\n", &TestCompositeExpr{Exprs: exprs, Op: OpOr})
	return exprs, nil
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
		rego.Input(DefaultInput),
		rego.Module(ModuleFileName, module),
	).PrepareForPartial(ctx)
	g.Expect(e).To(gomega.Succeed(), "preparing partial should not return error")

	pq, e := prepared.Partial(ctx)
	g.Expect(e).To(gomega.Succeed(), "partial evaluation should not return error")
	g.Expect(pq).ToNot(gomega.BeNil(), "partial queries should not be nil")
	return pq
}
