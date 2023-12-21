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
	"fmt"
	"github.com/onsi/gomega"
	"github.com/open-policy-agent/opa/ast"
	"sort"
	"strings"
)

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
	Compare(expr TestExpression) int
}

type TestExpr struct {
	Ref   string
	Value interface{}
	Op    string
}

func (e *TestExpr) Assert(_ context.Context, g *gomega.WithT, expected *ExpectedExpr) {
	g.Expect(expected.Exprs).To(gomega.HaveLen(0), "expect %s, but got comparison expression [%v]", expected.Op, e)
	g.Expect(e.Op).To(gomega.BeEquivalentTo(expected.Op), "op should be correct")
	g.Expect(e.Ref).To(gomega.BeEquivalentTo(expected.Ref), "ref should be correct")
	g.Expect(e.Value).To(gomega.BeEquivalentTo(expected.Value), "value should be correct")
}

func (e *TestExpr) Compare(expr TestExpression) int {
	switch v := expr.(type) {
	case *TestExpr:
		if ret := strings.Compare(e.Ref, v.Ref); ret != 0 {
			return ret
		}
		if ret := strings.Compare(e.Op, v.Op); ret != 0 {
			return ret
		}
		return strings.Compare(fmt.Sprint(e.Value), fmt.Sprint(v.Value))
	default:
		return -1
	}
}

func (e *TestExpr) String() string {
	return fmt.Sprintf("%s %s '%v'", e.Ref, e.Op, e.Value)
}

type TestCompositeExpr struct {
	Exprs []TestExpression
	Op    string
}

func (e *TestCompositeExpr) Assert(ctx context.Context, g *gomega.WithT, expected *ExpectedExpr) {
	g.Expect(e.Exprs).To(gomega.HaveLen(len(expected.Exprs)), "composite expressions should have correct length")
	g.Expect(e.Op).To(gomega.BeEquivalentTo(expected.Op), "op should be correct")
	for i := range e.Exprs {
		e.Exprs[i].Assert(ctx, g, expected.Exprs[i])
	}
}

func (e *TestCompositeExpr) Compare(expr TestExpression) int {
	switch v := expr.(type) {
	case *TestExpr:
		return 1
	case *TestCompositeExpr:
		if ret := len(e.Exprs) - len(v.Exprs); ret != 0 {
			return ret
		}
		if ret := strings.Compare(e.Op, v.Op); ret != 0 {
			return ret
		}
		return 0
	default:
		return -1
	}
}

func (e *TestCompositeExpr) String() (ret string) {
	strs := make([]string, len(e.Exprs))
	for i := range e.Exprs {
		strs[i] = fmt.Sprint(e.Exprs[i])
	}
	switch e.Op {
	case OpOr:
		ret = "(" + strings.Join(strs, " "+e.Op+" ") + ")"
	case OpAnd:
		ret = strings.Join(strs, " "+e.Op+" ")
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
	sort.SliceStable(exprs, func(i, j int) bool {
		return exprs[i].Compare(exprs[j]) < 0
	})
	return &TestCompositeExpr{
		Exprs: exprs,
		Op:    OpAnd,
	}
}

func (t TestQueryTranslator) Or(_ context.Context, exprs ...TestExpression) TestExpression {
	sort.SliceStable(exprs, func(i, j int) bool {
		return exprs[i].Compare(exprs[j]) < 0
	})
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

