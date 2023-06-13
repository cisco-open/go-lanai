package opadata

import (
	"context"
	"fmt"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

type PartialError struct {
	msg string
}

func (e PartialError) Error() string {
	return e.msg
}

func (e PartialError) Is(err error) bool {
	_, ok := err.(PartialError)
	return ok
}

func NewPartialError(tmpl string, args ...interface{}) error {
	return PartialError{
		msg: fmt.Sprintf(tmpl, args...),
	}
}

type illegalResolver struct{}

func (illegalResolver) Resolve(ast.Ref) (interface{}, error) {
	return nil, NewPartialError("resolving Ref is not supported")
}

type TranslateOptions[EXPR any]  func(opts *TranslateOption[EXPR])
type TranslateOption[EXPR any] struct {
	Translator QueryTranslator[EXPR]
}

func TranslatePartialQueries[EXPR any](ctx context.Context, pq *rego.PartialQueries, opts ...TranslateOptions[EXPR]) ([]EXPR, error) {
	opt := TranslateOption[EXPR]{}
	for _, fn := range opts {
		fn(&opt)
	}
	fmt.Printf("AST: %v\n", pq)
	exprs := make([]EXPR, 0, len(pq.Queries))
	for _, body := range pq.Queries {
		logger.WithContext(ctx).Debugf("Query: %v", body)
		ands := make([]EXPR, 0, 5)
		ast.WalkExprs(body, func(expr *ast.Expr) bool {
			qExpr, e := TranslateExpression(ctx, expr, &opt)
			if e != nil {
				return false
			}
			ands = append(ands, qExpr)
			return true
		})
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
		return zero, NewPartialError("unsupported Rego expression: %v", astExpr)
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
		err = NewPartialError("Unsupported Rego operation: %v", astExpr)
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
		return zero, NewPartialError(`invalid Rego operation format: expected "op(Ref, Value)", but got %v(%v)`, op, operands)
	}

	// resolve value
	value, e := ast.ValueToInterface(val, illegalResolver{})
	if e != nil {
		return zero, NewPartialError(`unable to resolve Rego value [%v]: %v`, val, e)
	}

	// resolve operator and column
	if ref.IsGround() {
		return opt.Translator.Comparison(ctx, op, ref, value)
	}

	ground := ref.GroundPrefix()
	op = OpIn
	return opt.Translator.Comparison(ctx, op, ground, value)
}


