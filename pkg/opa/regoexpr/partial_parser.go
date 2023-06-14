package regoexpr

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"fmt"
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
	if len(pq.Queries) == 0 {
		return nil, opa.QueriesNotResolvedError
	}
	opt := TranslateOption[EXPR]{}
	for _, fn := range opts {
		fn(&opt)
	}
	if opt.Translator == nil {
		return nil, ParsingError.WithMessage("query translator is nil")
	}

	fmt.Printf("AST: %v\n", pq)
	exprs := make([]EXPR, 0, len(pq.Queries))
	for _, body := range pq.Queries {
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

func TranslateExpression[EXPR any](ctx context.Context, astExpr *ast.Expr, opt *TranslateOption[EXPR]) (ret EXPR, err error) {
	//logger.WithContext(ctx).Debugf("Expr: %v", astExpr)
	//fmt.Printf("IsEquality = %v, IsGround = %v, IsCall = %v\n", astExpr.IsEquality(), astExpr.IsGround(), astExpr.IsCall())
	switch {
	case astExpr.OperatorTerm() != nil:
		ret, err = TranslateOperationExpr(ctx, astExpr, opt)
	case astExpr.IsEquality():
		ast.WalkTerms(astExpr, func(term *ast.Term) bool {
			fmt.Printf("Term : %T\n", term.Value)
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