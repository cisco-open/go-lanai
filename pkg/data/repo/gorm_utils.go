package repo

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

// index is used for Utility to specify index key.
// If index.Fields is set, index.Cols are ignored
type index []*schema.IndexOption

// GormUtils implements Utility interface
type GormUtils struct {
	api        GormApi
	meta       *GormMetadata
	uniqueness []index
}

func newGormUtils(api GormApi, meta *GormMetadata) GormUtils {
	indexes := meta.schema.ParseIndexes()
	uniqueness := make([]index, 0, len(indexes))
	for _, idx := range indexes {
		switch idx.Class {
		case "UNIQUE":
			fields := make([]*schema.IndexOption, len(idx.Fields))
			for i := range idx.Fields {
				fields[i] = &idx.Fields[i]
			}
			if len(fields) != 0 {
				uniqueness = append(uniqueness, fields)
			}
		}
	}
	return GormUtils{
		api:        api,
		meta:       meta,
		uniqueness: uniqueness,
	}
}

func (g GormUtils) CheckUniqueness(ctx context.Context, v interface{}, keys ...interface{}) (dups map[string]interface{}, err error) {
	if !g.meta.isSupportedValue(v, genericModelWrite) {
		return nil, ErrorInvalidUtilityUsage.
			WithMessage("%T is not a valid value for %s, requires %s", v, "CheckUniqueness", "*Struct, []*Struct, []Struct, or map[string]interface{}")
	}
	// index keys override
	var uniqueKeys []index
	if len(keys) == 0 {
		uniqueKeys = g.uniqueness
	} else if uniqueKeys, err = g.resolveIndexes(keys); err != nil {
		return nil, err
	}

	// where clause
	models := toInterfaces(v)
	var where clause.Where
	switch exprs := g.uniquenessWhere(models, uniqueKeys); {
	case len(exprs) == 0:
		return nil, ErrorInvalidUtilityUsage.WithMessage("%s is not possible. No non-zero unique fields found in given model [%v]", "CheckUniqueness", v)
	case len(exprs) == 1:
		where = clause.Where{Exprs: []clause.Expression{exprs[0]}}
	default:
		where = clause.Where{Exprs: []clause.Expression{clause.Or(exprs...)}}
	}

	// fetch and parse result
	existing := reflect.New(g.meta.schema.ModelType).Interface()
	switch rs := g.api.DB(ctx).Model(g.meta.model).Where(where).Limit(1).Take(existing); {
	case errors.Is(rs.Error, gorm.ErrRecordNotFound):
		return nil, nil
	case rs.Error != nil:
		return nil, rs.Error
	}

	// find duplicates
	for _, m := range models {
		dups = g.findDuplicateFields(m, existing, uniqueKeys)
		if len(dups) != 0 {
			break
		}
	}
	pairs := make([]string, 0, len(dups))
	for k, v := range dups {
		pairs = append(pairs, fmt.Sprintf("(%s)=(%v)", k, v))
	}
	return dups, data.NewDataError(data.ErrorCodeDuplicateKey, fmt.Errorf("duplicated values: %s", strings.Join(pairs, ", ")))
}

/************************
	Helpers
 ************************/

func (g GormUtils) resolveIndexes(keys []interface{}) ([]index, error) {
	ret := make([]index, 0, len(keys))
	for _, k := range keys {
		var idx index
		var e error
		switch v := k.(type) {
		case string:
			idx, e = g.asIndex([]string{v})
		case []string:
			idx, e = g.asIndex(v)
		default:
			return nil, ErrorInvalidUtilityUsage.WithMessage("Invalid key type %T", k)
		}
		if e != nil {
			return nil, e
		}
		ret = append(ret, idx)
	}
	return ret, nil
}

func (g GormUtils) asIndex(names []string) (index, error) {
	ret := make(index, len(names))
	for i, n := range names {
		f, paths := lookupField(g.meta.schema, n)
		switch {
		case f == nil:
			return nil, fmt.Errorf("field with name [%s] is not found on model %s", n, g.meta.schema)
		case len(paths) > 0:
			return nil, fmt.Errorf("associations are not supported in this utils")
		}
		ret[i] = &schema.IndexOption{ Field: f }
	}
	return ret, nil
}

func (g GormUtils) uniquenessWhere(models []interface{}, keys []index) (exprs []clause.Expression) {
	for _, m := range models {
		exprs = append(exprs, g.uniquenessExprs(reflect.ValueOf(m), keys)...)
	}
	return
}

func (g GormUtils) uniquenessExprs(modelV reflect.Value, keys []index) []clause.Expression {
	exprs := make([]clause.Expression, 0, len(keys))
	modelV = reflect.Indirect(modelV)
	for _, idx := range keys {
		if expr, ok := g.compositeEqExpr(modelV, idx); ok {
			exprs = append(exprs, expr)
		}
	}
	return exprs
}

// compositeEqExpr returns false if
// 1. all index values are zero values
// 2. any column value is not found
// 3. len(cols) == 0
func (g GormUtils) compositeEqExpr(modelV reflect.Value, idx index) (clause.Expression, bool) {
	allZero := true
	andExprs := make([]clause.Expression, len(idx))
	for i, f := range idx {
		v, ok := g.extractValue(modelV, f.Field)
		if !ok || !v.IsValid() {
			return nil, false
		}
		allZero = allZero && v.IsZero()
		andExprs[i] = clause.Eq{
			Column: clause.Column{Name: f.DBName},
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

// findDuplicateFields compare fields and returns fields that left and right are same
func (g GormUtils) findDuplicateFields(left interface{}, right interface{}, keys []index) map[string]interface{} {
	dups := map[string]interface{}{}
	leftV := reflect.Indirect(reflect.ValueOf(left))
	rightV := reflect.Indirect(reflect.ValueOf(right))
	for _, idx := range keys {
		for _, f := range idx {
			lVal, lok := g.extractValue(leftV, f.Field)
			if !lok {
				continue
			}
			rVal, rok := g.extractValue(rightV, f.Field)
			if !rok {
				continue
			}
			if reflect.DeepEqual(lVal.Interface(), rVal.Interface()) {
				dups[f.Name] = lVal.Interface()
			}
		}
	}
	return dups
}

func (g GormUtils) extractValue(modelV reflect.Value, f *schema.Field) (reflect.Value, bool) {
	switch modelV.Kind() {
	case reflect.Map:
		for i := modelV.MapRange(); i.Next(); {
			k, ok := i.Key().Interface().(string)
			if ok && (k == f.Name || k == f.DBName) {
				return i.Value(), true
			}
		}
	case reflect.Struct:
		return modelV.FieldByIndex(f.StructField.Index), true
	}
	return reflect.Value{}, false
}

func toInterfaces(v interface{}) (ret []interface{}) {
	rv := reflect.Indirect(reflect.ValueOf(v))
	switch rv.Kind() {
	case reflect.Slice:
		ret = make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			ret[i] = rv.Index(i).Interface()
		}
		return ret
	case reflect.Struct, reflect.Map:
		return []interface{}{v}
	}
	return nil
}