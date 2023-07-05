package regoexpr

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"reflect"
)

var logger = log.New("OPA.AST")

type TranslateOptions[EXPR any]  func(opts *TranslateOption[EXPR])
type TranslateOption[EXPR any] struct {
	Translator QueryTranslator[EXPR]
}

// TranslatePartialQueries translate OPA partial queries into other expression languages. e.g. Postgres expression
// Note:
// 1. When PartialQueries.Queries is empty, it means access is DENIED regardless any unknown values
// 2. When PartialQueries.Queries is not empty but contains nil body, it means access is GRANTED regardless any unknown values
func TranslatePartialQueries[EXPR any](ctx context.Context, pq *rego.PartialQueries, opts ...TranslateOptions[EXPR]) ([]EXPR, error) {
	logger.WithContext(ctx).Debugf("Queries: %v", pq)
	if len(pq.Queries) == 0 {
		return nil, opa.ErrQueriesNotResolved
	}

	opt := TranslateOption[EXPR]{}
	for _, fn := range opts {
		fn(&opt)
	}
	if opt.Translator == nil {
		return nil, ParsingError.WithMessage("query translator is nil")
	}

	queries := NormalizeQueries(ctx, pq.Queries)
	if len(queries) == 0 {
		return []EXPR{}, nil
	}

	exprs := make([]EXPR, 0, len(queries))
	for _, body := range queries {
		logger.WithContext(ctx).Debugf("Query: %v", body)
		ands := make([]EXPR, 0, 5)
		for _ ,expr := range body {
			if qExpr, e := TranslateExpression(ctx, expr, &opt); e != nil {
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
// TODO we should further normalizing queries by look at equalities:
// 		e.g. "value1 = input.resource.field AND value2 = input.resource.field" would always yield "false"
func NormalizeQueries(ctx context.Context, queries []ast.Body) []ast.Body {
	ret := make([]ast.Body, 0, len(queries))
	defer func() {
		if len(queries) != len(ret) {
			logger.WithContext(ctx).Debugf("Normalized Queries %d -> %d", len(queries), len(ret))
		}
	}()
	bodyHash := map[int]struct{}{}
	for _, body := range queries {
		if body == nil {
			// Because queries are "OR", if any query is nil, it means the entire queries always yield "true".
			// This means OPA can conclude the requested policy query without any unknowns
			return nil
		}

		// check duplicates
		if _, ok := bodyHash[body.Hash()]; ok {
			continue
		}
		bodyHash[body.Hash()] = struct{}{}

		// go through expressions
		exprs := make([]*ast.Expr, 0, len(body))
		exprHash := map[int]struct{}{}
		for _ ,expr := range body {
			if _, ok := exprHash[expr.Hash()]; !ok {
				exprHash[expr.Hash()] = struct{}{}
				exprs = append(exprs, expr)
			}
		}
		ret = append(ret, exprs)
	}
	return ret
}

func TranslateExpression[EXPR any](ctx context.Context, astExpr *ast.Expr, opt *TranslateOption[EXPR]) (ret EXPR, err error) {
	//logger.WithContext(ctx).Debugf("Expr: %v", astExpr)
	//fmt.Printf("IsEquality = %v, IsGround = %v, IsCall = %v\n", astExpr.IsEquality(), astExpr.IsGround(), astExpr.IsCall())
	switch {
	case astExpr.OperatorTerm() != nil:
		ret, err = TranslateOperationExpr(ctx, astExpr, opt)
	case astExpr.IsEquality():
		ast.WalkTerms(astExpr, func(term *ast.Term) bool {
			logger.WithContext(ctx).Debugf("Term: %T\n", term.Value)
			return true
		})
		var zero EXPR
		return zero, ParsingError.WithMessage("unsupported Rego expression: %v", astExpr)
	}
	return
}

func TranslateOperationExpr[EXPR any](ctx context.Context, astExpr *ast.Expr, opt *TranslateOption[EXPR]) (ret EXPR, err error) {
	op := astExpr.Operator()
	operands := astExpr.Operands()
	switch len(operands) {
	case 2:
		ret, err = TranslateThreeTermsOp(ctx, op, operands, opt)
	default:
		err = ParsingError.WithMessage("Unsupported Rego operation: %v", astExpr)
	}
	if err != nil {
		return
	}

	if astExpr.Negated {
		ret = opt.Translator.Negate(ctx, ret)
	}
	return
}

func TranslateThreeTermsOp[EXPR any](ctx context.Context, op ast.Ref, operands []*ast.Term, opt *TranslateOption[EXPR]) (ret EXPR, err error) {
	// format "op(Ref, Value)", "Ref op Value"
	var ref ast.Ref
	var val ast.Value
	var zero EXPR
	for _, term := range operands {
		switch v := term.Value.(type) {
		case ast.Ref:
			ref = v
		default:
			val = v
		}
	}
	if ref == nil || val == nil {
		return zero, ParsingError.WithMessage(`invalid Rego operation format: expected "op(Ref, Value)", but got %v(%v)`, op, operands)
	}

	// resolve value
	value, e := ast.ValueToInterface(val, illegalResolver{})
	if e != nil {
		return zero, ParsingError.WithMessage(`unable to resolve Rego value [%v]: %v`, val, e)
	}

	// resolve operator and column
	if ref.IsGround() {
		return opt.Translator.Comparison(ctx, op, ref, value)
	}

	ground := ref.GroundPrefix()
	op = OpIn
	return opt.Translator.Comparison(ctx, op, ground, value)
}

type illegalResolver struct{}

func (illegalResolver) Resolve(ast.Ref) (interface{}, error) {
	return nil, ParsingError.WithMessage("resolving Ref is not supported")
}