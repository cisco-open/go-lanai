package opadata

import (
	"context"
	"github.com/open-policy-agent/opa/ast"
	"gorm.io/gorm/clause"
)

type GormTranslator struct {

}

func (t *GormTranslator) Quote(ctx context.Context, field interface{}) string {
	//TODO implement me
	panic("implement me")
}

func (t *GormTranslator) Resolve(ctx context.Context, colRef ast.Ref) (clause.Column, error) {
	//TODO implement me
	panic("implement me")
}

func (t *GormTranslator) ComparisonTranslator(ctx context.Context, op ast.Ref, col, val interface{}) (clause.Expression, error) {
	//TODO implement me
	panic("implement me")
}
