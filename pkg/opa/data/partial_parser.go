package opadata

import (
	"context"
	"fmt"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"gorm.io/gorm/clause"
	"strings"
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

func ParsePartialQueries(ctx context.Context, pq *rego.PartialQueries) ([]clause.Expression, error) {
	fmt.Printf("AST: %v\n", pq)
	exprs := make([]clause.Expression, 0, len(pq.Queries))
	for _, body := range pq.Queries {
		logger.WithContext(ctx).Debugf("Query: %v", body)
		ands := make([]clause.Expression, 0, 5)
		ast.WalkExprs(body, func(expr *ast.Expr) bool {
			gormExpr, e := TranslateExpression(ctx, expr)
			if e != nil {
				return false
			}
			ands = append(ands, gormExpr)
			return true
		})
		exprs = append(exprs, clause.And(ands...))
	}
	return exprs, nil
}

func TranslateExpression(ctx context.Context, astExpr *ast.Expr) (ret clause.Expression, err error) {
	//logger.WithContext(ctx).Debugf("Expr: %v", astExpr)
	//fmt.Printf("IsEquality = %v, IsGround = %v, IsCall = %v\n", astExpr.IsEquality(), astExpr.IsGround(), astExpr.IsCall())
	switch {
	case astExpr.OperatorTerm() != nil:
		ret, err = TranslateOperationExpr(ctx, astExpr)
	case astExpr.IsEquality():
		ast.WalkTerms(astExpr, func(term *ast.Term) bool {
			fmt.Printf("Term : %T\n", term.Value)
			return true
		})
		return nil, NewPartialError("unsupported Rego expression: %v", astExpr)
	}
	return
}

func TranslateOperationExpr(ctx context.Context, astExpr *ast.Expr) (ret clause.Expression, err error) {
	op := astExpr.Operator()
	operands := astExpr.Operands()
	switch len(operands) {
	case 2:
		ret, err = TranslateThreeTermsOp(ctx, op, operands)
	default:
		return nil, NewPartialError("Unsupported Rego operation: %v", astExpr)
	}
	if astExpr.Negated {
		ret = clause.Not(ret)
	}
	return
}

func TranslateThreeTermsOp(ctx context.Context, op ast.Ref, operands []*ast.Term) (ret clause.Expression, err error) {
	// format "op(Ref, Value)", "Ref op Value"
	var ref ast.Ref
	var val ast.Value
	for _, term := range operands {
		switch v := term.Value.(type) {
		case ast.Ref:
			ref = v
		default:
			val = v
		}
	}
	if ref == nil || val == nil {
		return nil, NewPartialError(`invalid Rego operation format: expected "op(Ref, Value)", but got %v(%v)`, op, operands)
	}

	// resolve value
	value, e := ast.ValueToInterface(val, illegalResolver{})
	if e != nil {
		return nil, NewPartialError(`unable to resolve Rego value [%v]: %v`, val, e)
	}

	// resolve operator and column
	if ref.IsGround() {
		col, e := ToColumn(ctx, ref)
		if e != nil {
			return nil, NewPartialError(`unable to column name from Rego reference [%v]: %v`, ref, e)
		}
		return ToColumnComparison(ctx, op, col, value)
	}

	ground := ref.GroundPrefix()
	op = OpIn
	return ToColumnComparison(ctx, op, ground, value)
}

var (
	TermInternal = ast.VarTerm("internal")
)

var (
	OpIn    ast.Ref = []*ast.Term{TermInternal, ast.StringTerm("in")}
	OpEqual         = ast.Equality.Ref()
	OpEq            = ast.Equal.Ref()
	OpNeq           = ast.NotEqual.Ref()
	OpLte           = ast.LessThanEq.Ref()
	OpLt            = ast.LessThan.Ref()
	OpGte           = ast.GreaterThanEq.Ref()
	OpGt            = ast.GreaterThan.Ref()
)

var (
	OpHashEqual = OpEqual.Hash()
	OpHashEq    = OpEq.Hash()
	OpHashNeq   = OpNeq.Hash()
	OpHashLte   = OpLte.Hash()
	OpHashLt    = OpLt.Hash()
	OpHashGte   = OpGte.Hash()
	OpHashGt    = OpGt.Hash()
	OpHashIn    = OpIn.Hash()
)

func ToColumn(_ context.Context, colRef ast.Ref) (interface{}, error) {
	// TODO
	path := colRef.String()
	idx := strings.LastIndex(path, ".")
	return clause.Column{
		Table: path[idx+1:],
		Name:  path[:idx],
	}, nil
}

func ToColumnComparison(_ context.Context, op ast.Ref, col, val interface{}) (ret clause.Expression, err error) {
	switch op.Hash() {
	case OpHashEqual, OpHashEq:
		ret = &clause.Eq{Column: col, Value: val}
	case OpHashNeq:
		ret = &clause.Neq{Column: col, Value: val}
	case OpHashLte:
		ret = &clause.Lte{Column: col, Value: val}
	case OpHashLt:
		ret = &clause.Lt{Column: col, Value: val}
	case OpHashGte:
		ret = &clause.Gte{Column: col, Value: val}
	case OpHashGt:
		ret = &clause.Gt{Column: col, Value: val}
	case OpHashIn:
		colSql := col // TODO should use statement quote
		sql := fmt.Sprintf("%s @> ?", colSql)
		ret = clause.Expr{
			SQL:  sql,
			Vars: []interface{}{[]interface{}{val}},
		}
	default:
		return nil, NewPartialError("Unsupported Rego operator: %v", op)
	}
	return
}
