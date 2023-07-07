package regoexpr

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"github.com/open-policy-agent/opa/ast"
)

var (
	ParsingError = opa.NewError(`generic OPA partial query parsing error`)
)

type QueryTranslator[EXPR any] interface {
	Negate(ctx context.Context, expr EXPR) EXPR
	And(ctx context.Context, expr ...EXPR) EXPR
	Or(ctx context.Context, expr ...EXPR) EXPR
	Comparison(ctx context.Context, op ast.Ref, colRef ast.Ref, val interface{}) (EXPR, error)
}

var (
	TermInternal = ast.VarTerm("internal")
	OpInternal   = ast.Ref([]*ast.Term{TermInternal})
	OpIn         = ast.Ref([]*ast.Term{TermInternal, ast.StringTerm("in")})
	OpEqual      = ast.Equality.Ref()
	OpEq         = ast.Equal.Ref()
	OpNeq        = ast.NotEqual.Ref()
	OpLte        = ast.LessThanEq.Ref()
	OpLt         = ast.LessThan.Ref()
	OpGte        = ast.GreaterThanEq.Ref()
	OpGt         = ast.GreaterThan.Ref()
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
