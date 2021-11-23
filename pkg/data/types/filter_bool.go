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
	"strings"
)

const (
	TagFilter = "filter"
)

/****************************
	Func
 ****************************/

// SkipBoolFilter is a gorm scope that can be used to skip filtering of FilterBool fields with given field names
// e.g. db.WithContext(ctx).Scopes(SkipBoolFilter("FieldName1", "FieldName2")).Find(...)
//
// To disable all FilterBool filtering, provide no params or "*"
// e.g. db.WithContext(ctx).Scopes(SkipBoolFilter()).Find(...)
//
// Note using this scope without context would panic
func SkipBoolFilter(filedNames ...string) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if tx.Statement.Context == nil {
			panic("SkipBoolFilter used without context")
		}
		ctx := tx.Statement.Context
		for _, filedName := range filedNames {
			ctx = context.WithValue(ctx, ckFilterMode(filedName), fmDisabled)
		}
		if len(filedNames) == 0 {
			ctx = context.WithValue(ctx, ckFilterMode("*"), fmDisabled)
		}
		tx.Statement.Context = ctx
		return tx
	}
}

// BoolFiltering is a gorm scope that change default/tag behavior of FilterBool field filtering with given field names
// e.g. db.WithContext(ctx).Scopes(BoolFiltering(false, "FieldName1", "FieldName2")).Find(...)
// 		would filter out any model with "FieldName1" or "FieldName2" equals to "false"
//
// To override all FilterBool filtering, provide no params or "*"
// e.g. db.WithContext(ctx).Scopes(BoolFiltering()).Find(...)
//
// Note using this scope without context would panic
func BoolFiltering(filterVal bool, filedNames ...string) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if tx.Statement.Context == nil {
			panic("BoolFiltering used without context")
		}
		mode := fmPositive
		if !filterVal {
			mode = fmNegative
		}
		ctx := tx.Statement.Context
		for _, filedName := range filedNames {
			ctx = context.WithValue(ctx, ckFilterMode(filedName), mode)
		}
		if len(filedNames) == 0 {
			ctx = context.WithValue(ctx, ckFilterMode("*"), mode)
		}
		tx.Statement.Context = ctx
		return tx
	}
}

/****************************
	Types
 ****************************/

type ckFilterMode string

const (
	fmPositive filterMode = iota
	fmNegative
	fmDisabled
)

// filterMode enum of possible values fm*
type filterMode int

// FilterBool implements
// - schema.GormDataTypeInterface
// - schema.QueryClausesInterface
// this data type adds "WHERE" clause in SELECT statements for filtering out models based on this field's value
//
// FilterBool by default filter out true values (WHERE filter_bool_col IS NOT TRUE AND ....).
// this behavior can be changed to using tag `filter:"<-|true|false>"`
// - `filter:"-"`: 		disables the filtering at model declaration level.
//						Can be enabled on per query basis using scopes or repo options (if applicable)
// - `filter:"true"`: 	filter out "true" values, the default behavior
//						Can be overridden on per query basis using scopes or repo options (if applicable)
// - `filter:"false"`: 	filter out "false" values
//						Can be overridden on per query basis using scopes or repo options (if applicable)
// See SkipBoolFilter and BoolFiltering for filtering behaviour overriding
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
	return []clause.Interface{newBoolFilterClause(f)}
}

/****************************
	Helpers
 ****************************/

// boolFilterClause implements clause.Interface and gorm.StatementModifier, where gorm.StatementModifier do the real work.
// See gorm.DeletedAt for impl. reference
type boolFilterClause struct {
	stmtModifier
	FilterMode filterMode
	Field      *schema.Field
}

func newBoolFilterClause(f *schema.Field) clause.Interface {
	mode := fmPositive
	tag := strings.ToLower(strings.TrimSpace(f.Tag.Get(TagFilter)))
	switch tag {
	case "false":
		mode = fmNegative
	case "-":
		mode = fmDisabled
	}

	return &boolFilterClause{
		FilterMode: mode,
		Field:      f,
	}
}

func (c boolFilterClause) ModifyStatement(stmt *gorm.Statement) {
	mode := c.determineFilterMode(stmt.Context)
	if mode == fmDisabled {
		return
	}

	// special fix for db.Model(&model{}).Where(&model{f1:v1}).Or(&model{f2:v2})...
	// Ref:	https://github.com/go-gorm/gorm/issues/3627
	//		https://github.com/go-gorm/gorm/commit/9b2181199d88ed6f74650d73fa9d20264dd134c0#diff-e3e9193af67f3a706b3fe042a9f121d3609721da110f6a585cdb1d1660fd5a3c
	fixWhereClausesForStatementModifier(stmt)

	// add bool filtering
	colExpr := stmt.Quote(clause.Column{Table: clause.CurrentTable, Name: c.Field.DBName})
	unfilteredValue := mode != fmPositive
	stmt.AddClause(clause.Where{Exprs: []clause.Expression{
		clause.Expr{
			SQL: fmt.Sprintf("%s IS %s", colExpr, strconv.FormatBool(unfilteredValue)),
		},
	}})
}

/***********************
	Helpers
 ***********************/

func (c boolFilterClause) determineFilterMode(ctx context.Context) filterMode {
	if ctx == nil {
		return c.FilterMode
	}
	if v, ok := ctx.Value(ckFilterMode("*")).(filterMode); ok {
		return v
	}
	if v, ok := ctx.Value(ckFilterMode(c.Field.Name)).(filterMode); ok {
		return v
	}
	return c.FilterMode
}
