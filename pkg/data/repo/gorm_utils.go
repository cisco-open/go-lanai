package repo

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
)

// compositeKey is used for Utility to specify composite key.
// If compositeKey.Fields is set, compositeKey.Cols are ignored
type compositeKey []*clause.Column

// GormUtils implements Utility interface
type GormUtils struct {
	api        GormApi
	meta       *GormMetadata
	uniqueness []compositeKey
}

func newGormUtils(api GormApi, meta *GormMetadata) GormUtils {
	indexes := meta.schema.ParseIndexes()
	uniqueness := make([]compositeKey, 0, len(indexes))
	for _, idx := range indexes {
		switch idx.Class {
		case "UNIQUE":
			cols := make([]*clause.Column, len(idx.Fields))
			for i, f := range idx.Fields {
				cols[i] = &clause.Column{ Name:  f.DBName }
			}
			if len(cols) != 0 {
				uniqueness = append(uniqueness, cols)
			}
		}
	}
	return GormUtils{
		api:        api,
		meta:       meta,
		uniqueness: uniqueness,
	}
}

func (g GormUtils) CheckUniqueness(ctx context.Context, v interface{}, keys ...interface{}) (err error) {
	if !g.meta.isSupportedValue(v, genericModelWrite) {
		return ErrorInvalidUtilityUsage.
			WithMessage("%T is not a valid value for %s, requires %s", v, "CheckUniqueness", "*Struct, []*Struct, []Struct, or map[string]interface{}")
	}
	var uniqueKeys []compositeKey
	if len(keys) == 0 {
		uniqueKeys = g.uniqueness
	} else if uniqueKeys, err = g.prepareKeys(keys); err != nil {
		return err
	}

	var where clause.Where
	exprs, e := g.uniquenessWhere(reflect.ValueOf(v), uniqueKeys)
	switch {
	case e != nil:
		return e
	case len(exprs) == 0:
		return ErrorInvalidUtilityUsage.WithMessage("%s is not possible. No non-zero unique fields found in given model [%v]", "CheckUniqueness", v)
	case len(exprs) == 1:
		where = clause.Where{Exprs: []clause.Expression{exprs[0]}}
	default:
		where = clause.Where{Exprs: []clause.Expression{clause.Or(exprs...)}}
	}
	var count int64

	switch e := execute(ctx, g.api.DB(ctx), nil, nil, modelFunc(g.meta.model), func(db *gorm.DB) *gorm.DB {
		return db.Where(where).Count(&count)
	}); {
	case e != nil:
		return e
	case count > 0:
		return data.NewDataError(data.ErrorCodeDuplicateKey, "duplicated key")
	default:
		return nil
	}
}

/************************
	Helpers
 ************************/

func (g GormUtils) prepareKeys(keys []interface{}) ([]compositeKey, error) {
	ret := make([]compositeKey, 0, len(keys))
	for _, k := range keys {
		switch v := k.(type) {
		case string:
			if col, e := toColumn(g.meta.schema, v); e == nil {
				ret = append(ret, compositeKey{col})
			}
		case []string:
			ck := make(compositeKey, 0, len(v))
			for _, f := range v {
				col, e := toColumn(g.meta.schema, f)
				if e != nil {
					return nil, ErrorInvalidUtilityUsage.WithMessage("Invalid key %v", f)
				}
				ck = append(ck, col)
			}
			if len(ck) != 0 {
				ret = append(ret, ck)
			}
		default:
			return nil, ErrorInvalidUtilityUsage.WithMessage("Invalid key type %T", k)
		}
	}
	return ret, nil
}

func (g GormUtils) uniquenessWhere(v reflect.Value, keys []compositeKey) (exprs []clause.Expression, err error) {
	rv := reflect.Indirect(v)
	switch rv.Kind() {
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			mv := rv.Index(i)
			sub, e := g.uniquenessWhere(mv, keys)
			if e != nil {
				return nil, e
			}
			exprs = append(exprs, sub...)
		}
	case reflect.Struct:
		return g.structUniquenessExprs(rv, keys), nil
	case reflect.Map:
		return g.mapUniquenessExprs(rv, keys), nil
	default:
		return nil, ErrorInvalidUtilityUsage
	}
	return
}

func (g GormUtils) structUniquenessExprs(modelV reflect.Value, keys []compositeKey) []clause.Expression {
	exprs := make([]clause.Expression, 0, len(keys))
	modelV = reflect.Indirect(modelV)
	for _, ck := range keys {
		if expr, ok := g.compositeEqExpr(modelV, ck); ok {
			exprs = append(exprs, expr)
		}
	}
	return exprs
}

func (g GormUtils) mapUniquenessExprs(m reflect.Value, keys []compositeKey) []clause.Expression {
	exprs := make([]clause.Expression, 0, len(keys))
	for _, ck := range keys {
		if expr, ok := g.compositeEqExpr(m, ck); ok {
			exprs = append(exprs, expr)
		}
	}
	return exprs
}

// compositeEqExpr returns false if
// 1. all composite values are zero values
// 2. any column value is not found
// 3. len(cols) == 0
func (g GormUtils) compositeEqExpr(modelV reflect.Value, cols compositeKey) (clause.Expression, bool) {
	allZero := true
	andExprs := make([]clause.Expression, len(cols))
	for i, col := range cols {
		v, ok := g.extractValue(modelV, col.Name)
		if !ok || !v.IsValid() {
			return nil, false
		}
		allZero = allZero && v.IsZero()
		andExprs[i] = clause.Eq{
			Column: clause.Column{Name: col.Name},
			Value:  v.Interface(),
		}
	}
	switch {
	case allZero || len(andExprs) == 0:
		return nil, false
	case len(andExprs) == 1:
		return andExprs[0], true
	default:
		return clause.And(andExprs...), true
	}
}

func (g GormUtils) extractValue(modelV reflect.Value, col string) (reflect.Value, bool) {
	switch modelV.Kind() {
	case reflect.Map:
		for i := modelV.MapRange(); i.Next(); {
			k, ok := i.Key().Interface().(string)
			if ok && (k == col || g.meta.ColumnName(k) == col) {
				return i.Value(), true
			}
		}
	case reflect.Struct:
		f, ok := g.meta.schema.FieldsByDBName[col]
		if !ok {
			return reflect.Value{}, false
		}
		return modelV.FieldByIndex(f.StructField.Index), true
	}
	return reflect.Value{}, false
}
