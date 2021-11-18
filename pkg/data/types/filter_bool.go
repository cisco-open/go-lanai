package types

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// BoolFilter implements
// - schema.GormDataTypeInterface
// - schema.QueryClausesInterface
// this data type adds "WHERE" clause in SELECT statements for filtering models with this field == true
type BoolFilter bool

// Value implements driver.Valuer
func (t BoolFilter) Value() (driver.Value, error) {
	return sql.NullBool{
		Bool:  bool(t),
		Valid: true,
	}.Value()
}

// Scan implements sql.Scanner
func (t *BoolFilter) Scan(src interface{}) error {
	nullBool := &sql.NullBool{}
	if e := nullBool.Scan(src); e != nil {
		return e
	}
	*t = BoolFilter(nullBool.Valid && nullBool.Bool)
	return nil
}

func (t BoolFilter) GormDataType() string {
	return "bool"
}

// QueryClauses implements schema.QueryClausesInterface,
func (t BoolFilter) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newBoolFilterClause(f, true)}
}

// tenancyFilterClause implements clause.Interface and gorm.StatementModifier, where gorm.StatementModifier do the real work.
// See gorm.DeletedAt for impl. reference
type boolFilterClause struct {
	FilteredValue bool
	Field *schema.Field
}

func newBoolFilterClause(f *schema.Field, filterValue bool) clause.Interface {
	return &boolFilterClause{
		FilteredValue: filterValue,
		Field:         f,
	}
}

func (c boolFilterClause) Name() string {
	return ""
}

func (c boolFilterClause) Build(clause.Builder) {
}

func (c boolFilterClause) MergeClause(*clause.Clause) {
}

func (c boolFilterClause) ModifyStatement(stmt *gorm.Statement) {
	if shouldSkipBoolFilter(stmt.Context) {
		return
	}

	// special fix for db.Model(&model{}).Where(&model{f1:v1}).Or(&model{f2:v2})...
	// Ref:	https://github.com/go-gorm/gorm/issues/3627
	//		https://github.com/go-gorm/gorm/commit/9b2181199d88ed6f74650d73fa9d20264dd134c0#diff-e3e9193af67f3a706b3fe042a9f121d3609721da110f6a585cdb1d1660fd5a3c
	fixWhereClausesForStatementModifier(stmt)

	// add bool filtering
	stmt.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Neq{
			Column: clause.Column{ Table: clause.CurrentTable, Name:  c.Field.DBName },
			Value:  c.FilteredValue,
		},
	}})
}

/***********************
	Helpers
 ***********************/

func shouldSkipBoolFilter(ctx context.Context) bool {
	//return ctx == nil || ctx.Value(ckSkipTenancyCheck{}) != nil || !security.IsFullyAuthenticated(security.Get(ctx))
	return false
}


