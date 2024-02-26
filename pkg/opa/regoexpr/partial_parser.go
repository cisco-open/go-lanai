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
    "encoding/json"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/cisco-open/go-lanai/pkg/opa"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/open-policy-agent/opa/ast"
    "github.com/open-policy-agent/opa/rego"
    "reflect"
)

var logger = log.New("OPA.AST")

type TranslateOptions[EXPR any] func(opts *TranslateOption[EXPR])
type TranslateOption[EXPR any] struct {
	Translator QueryTranslator[EXPR]
}

// TranslatePartialQueries translate OPA partial queries into other expression languages. e.g. Postgres expression
// Note:
// 1. When PartialQueries.Queries is empty, it means access is DENIED regardless any unknown values
// 2. When PartialQueries.Queries is not empty but contains nil body, it means access is GRANTED regardless any unknown values
func TranslatePartialQueries[EXPR any](ctx context.Context, pq *rego.PartialQueries, opts ...TranslateOptions[EXPR]) ([]EXPR, error) {
	logger.WithContext(ctx).Debugf("Queries: %v", pq)

	opt := TranslateOption[EXPR]{}
	for _, fn := range opts {
		fn(&opt)
	}
	if opt.Translator == nil {
		return nil, ParsingError.WithMessage("query translator is nil")
	}

	// normalize
	queries, changed := NormalizeQueries(ctx, pq.Queries)
	if changed {
		logger.WithContext(ctx).Debugf("Normalized Queries: %v", queries)
	}
	// If queries is nil, it means any unknowns can satisfy.
	// However, if queries is empty, it means no unknowns would satisfy
	switch {
	case queries == nil:
		return []EXPR{}, nil
	case len(queries) == 0:
		return nil, opa.ErrQueriesNotResolved
	}

	exprs := make([]EXPR, 0, len(queries))
	for _, body := range queries {
		logger.WithContext(ctx).Debugf("Parsing Query: %v", body)
		ands := make([]EXPR, 0, 5)
		for _, expr := range body {
			if qExpr, e := TranslateExpression(ctx, expr, &opt); e != nil {
				logger.WithContext(ctx).Debugf("%v", e)
				return nil, e
			} else if !reflect.ValueOf(qExpr).IsZero() {
				ands = append(ands, qExpr)
			}
		}
		exprs = append(exprs, opt.Translator.And(ctx, ands...))
	}
	return exprs, nil
}

// NormalizeQueries remove duplicate queries and duplicate expressions in each query
func NormalizeQueries(ctx context.Context, queries []ast.Body) (ret []ast.Body, changed bool) {
	ret = make([]ast.Body, 0, len(queries))
	bodyHash := map[int]struct{}{}
	for _, body := range queries {
		if body == nil {
			// Because queries are "OR", if any query is nil, it means the entire queries always yield "true".
			// This means OPA can conclude the requested policy query without any unknowns
			return nil, true
		}

		// check duplicates
		if _, ok := bodyHash[body.Hash()]; ok {
			changed = true
			continue
		}
		bodyHash[body.Hash()] = struct{}{}

		// normalize body
		if exprs, ok := NormalizeExpressions(ctx, body); len(exprs) != 0 {
			ret = append(ret, exprs)
			changed = changed || ok
		} else {
			changed = true
		}
	}
	return
}

// NormalizeExpressions remove duplicate expressions in query
func NormalizeExpressions(ctx context.Context, body ast.Body) (exprs ast.Body, changed bool) {
	exprs = make([]*ast.Expr, 0, len(body))
	if HasControversialExpressions(ctx, body) {
		logger.WithContext(ctx).Debugf("Controversial Query: %v", body)
		return exprs, true
	}
	exprHash := map[int]struct{}{}
	for _, expr := range body {
		hash := calculateHash(expr)
		if _, ok := exprHash[hash]; !ok {
			exprHash[hash] = struct{}{}
			exprs = append(exprs, expr)
		} else {
			changed = true
		}
	}
	return
}

// HasControversialExpressions analyze given expression and return true if it contains controversial expressions:
// Examples:
// - "value1 = input.resource.field AND value2 = input.resource.field"
// - "value1 = input.resource.field AND value1 != input.resource.field"
func HasControversialExpressions(_ context.Context, body []*ast.Expr) (ret bool) {
	equals := map[int]int{}
	notEquals := map[int]utils.Set{}
	for _, expr := range body {
		ref, val, op, ok := resolveThreeTermsOp(expr)
		if !ok || !ref.IsGround() {
			continue
		}
		// only handle equalities
		negate := false
		switch {
		case OpEqual.Equal(op) || OpEq.Equal(op):
			negate = expr.Negated
		case OpNeq.Equal(op):
			negate = !expr.Negated
		default:
			continue
		}
		// compare values by hashes
		rHash := ref.Hash()
		vHash := val.Hash()
		if negate {
			if _, ok := notEquals[rHash]; !ok {
				notEquals[rHash] = utils.NewSet()
			}
			notEquals[rHash].Add(vHash)
		} else {
			// var == v1 AND var == v2
			if v, ok := equals[rHash]; ok && v != vHash {
				return true
			}
			equals[rHash] = vHash
		}
		// var == v1 AND var != v1
		if notEquals[rHash].Has(equals[rHash]) {
			return true
		}
	}
	return false
}

func TranslateExpression[EXPR any](ctx context.Context, astExpr *ast.Expr, opt *TranslateOption[EXPR]) (ret EXPR, err error) {
	//logger.WithContext(ctx).Debugf("Expr: %v", astExpr)
	//fmt.Printf("IsEquality = %v, IsGround = %v, IsCall = %v\n", astExpr.IsEquality(), astExpr.IsGround(), astExpr.IsCall())
	switch {
	case astExpr.OperatorTerm() != nil:
		ret, err = TranslateOperationExpr(ctx, astExpr, opt)
	default:
		return ret, ParsingError.WithMessage("unsupported Rego expression: %v", astExpr)
	}
	return
}

func TranslateOperationExpr[EXPR any](ctx context.Context, astExpr *ast.Expr, opt *TranslateOption[EXPR]) (ret EXPR, err error) {
	operands := astExpr.Operands()
	switch len(operands) {
	case 2:
		ret, err = TranslateThreeTermsOp(ctx, astExpr, opt)
	default:
		err = ParsingError.WithMessage("unsupported Rego operation: %v", astExpr)
	}
	if err != nil {
		return
	}

	if astExpr.Negated {
		ret = opt.Translator.Negate(ctx, ret)
	}
	return
}

func TranslateThreeTermsOp[EXPR any](ctx context.Context, astExpr *ast.Expr, opt *TranslateOption[EXPR]) (ret EXPR, err error) {
	// format "op(Ref, Value)", "Ref op Value"
	var zero EXPR
	ref, val, op, ok := resolveThreeTermsOp(astExpr)
	switch {
	case !ok:
		return zero, ParsingError.WithMessage(`invalid Rego operation format: expected "op(Ref, Value)", but got %v(%v)`, astExpr.OperatorTerm(), astExpr.Operands())
	case op.HasPrefix(OpInternal):
		return zero, ParsingError.WithMessage(`unsupported Rego operator [%v]`, op)
	}

	// resolve value
	value, e := ast.ValueToInterface(val, illegalResolver{})
	if e != nil {
		return zero, ParsingError.WithMessage(`unable to resolve Rego value [%v]: %v`, val, e)
	}
	if v, ok := value.(json.Number); ok {
		if value, e = v.Float64(); e != nil {
			return zero, ParsingError.WithMessage(`unable to resolve Rego value [%v] as number: %v`, val, e)
		}
	}

	// resolve operator and column
	if ref.IsGround() {
		return opt.Translator.Comparison(ctx, op, ref, value)
	}

	ground := ref.GroundPrefix()
	op = OpIn
	return opt.Translator.Comparison(ctx, op, ground, value)
}

/**********************
	Helpers
 **********************/

// note: when we calculate hash, we don't want to consider its Index and non-ground part
func calculateHash(astExpr *ast.Expr) int {
	expr := astExpr.Copy()
	expr.Index = 0
	return expr.Hash()
}

func resolveThreeTermsOp(astExpr *ast.Expr) (ref ast.Ref, val ast.Value, op ast.Ref, ok bool) {
	// format "op(Ref, Value)", "Ref op Value"
	op = astExpr.Operator()
	operands := astExpr.Operands()
	if op == nil || len(operands) != 2 {
		return nil, nil, nil, false
	}

	for _, term := range operands {
		switch v := term.Value.(type) {
		case ast.Ref:
			ref = v
		default:
			val = v
		}
	}
	if ref == nil || val == nil {
		return nil, nil, nil, false
	}
	return ref, val, op, true
}

type illegalResolver struct{}

func (illegalResolver) Resolve(ast.Ref) (interface{}, error) {
	return nil, ParsingError.WithMessage("resolving Ref is not supported")
}
