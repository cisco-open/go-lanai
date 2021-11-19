package types

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"strconv"
)

/****************************
	Func
 ****************************/

// SkipBoolFilter is a gorm scope that can be used to skip filtering of FilterBool and NegFilterBool
// e.g. db.WithContext(ctx).Scopes(SkipBoolFilter()).Find(...)
// Note using this scope without context would panic
func SkipBoolFilter() func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if tx.Statement.Context == nil {
			panic("SkipBoolFilter used without context")
		}
		ctx := context.WithValue(tx.Statement.Context, ckSkipBoolFilter{}, struct{}{})
		tx.Statement.Context = ctx
		return tx
	}
}

/****************************
	Types
 ****************************/

type ckSkipBoolFilter struct{}

// FilterBool implements
// - schema.GormDataTypeInterface
// - schema.QueryClausesInterface
// this data type adds "WHERE" clause in SELECT statements for filtering out models if this field == true
type FilterBool bool

// Value implements driver.Valuer
func (t FilterBool) Value() (driver.Value, error) {
	return sql.NullBool{
		Bool:  bool(t),
		Valid: true,
	}.Value()
}

// Scan implements sql.Scanner
func (t *FilterBool) Scan(src interface{}) error {
	nullBool := &sql.NullBool{}
	if e := nullBool.Scan(src); e != nil {
		return e
	}
	*t = FilterBool(nullBool.Valid && nullBool.Bool)
	return nil
}

func (t FilterBool) GormDataType() string {
	return "bool"
}

// QueryClauses implements schema.QueryClausesInterface,
func (t FilterBool) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newBoolFilterClause(f, true)}
}

// NegFilterBool implements
// - schema.GormDataTypeInterface
// - schema.QueryClausesInterface
// this data type adds "WHERE" clause in SELECT statements for filtering out models if this field == false
type NegFilterBool bool

// Value implements driver.Valuer
func (t NegFilterBool) Value() (driver.Value, error) {
	return sql.NullBool{
		Bool:  bool(t),
		Valid: true,
	}.Value()
}

// Scan implements sql.Scanner
func (t *NegFilterBool) Scan(src interface{}) error {
	nullBool := &sql.NullBool{}
	if e := nullBool.Scan(src); e != nil {
		return e
	}
	*t = NegFilterBool(nullBool.Valid && nullBool.Bool)
	return nil
}

func (t NegFilterBool) GormDataType() string {
	return "bool"
}

// QueryClauses implements schema.QueryClausesInterface,
func (t NegFilterBool) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newBoolFilterClause(f, false)}
}

/****************************
	Helpers
 ****************************/

// boolFilterClause implements clause.Interface and gorm.StatementModifier, where gorm.StatementModifier do the real work.
// See gorm.DeletedAt for impl. reference
type boolFilterClause struct {
	stmtModifier
	FilteredValue bool
	Field         *schema.Field
}

func newBoolFilterClause(f *schema.Field, filterValue bool) clause.Interface {
	return &boolFilterClause{
		FilteredValue: filterValue,
		Field:         f,
	}
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
	colExpr := stmt.Quote(clause.Column{Table: clause.CurrentTable, Name: c.Field.DBName})
	stmt.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Expr{
			SQL: fmt.Sprintf("%s IS %s", colExpr, strconv.FormatBool(!c.FilteredValue)),
		},
	}})
}

/***********************
	Helpers
 ***********************/

func shouldSkipBoolFilter(ctx context.Context) bool {
	return ctx != nil && ctx.Value(ckSkipBoolFilter{}) != nil
}
