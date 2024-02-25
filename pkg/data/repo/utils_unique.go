// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package repo

import (
    "context"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/data"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
    "gorm.io/gorm/schema"
    "reflect"
    "sort"
    "strings"
)

// index is used for Utility to specify index key.
// If index.Fields is set, index.Cols are ignored
type index []*schema.IndexOption

func gormCheckUniqueness(ctx context.Context, g GormApi, resolver GormSchemaResolver, v interface{}, keys []interface{}) (dups map[string]interface{}, err error) {
	// index keys override
	var uniqueKeys []index
	if len(keys) == 0 {
		uniqueKeys = findUniqueIndexes(resolver.Schema())
	} else if uniqueKeys, err = resolveIndexes(resolver.Schema(), keys); err != nil {
		return nil, err
	}

	// where clause
	models := toInterfaces(v)
	var where clause.Where
	switch exprs := uniquenessWhere(models, uniqueKeys); {
	case len(exprs) == 0:
		return nil, ErrorInvalidUtilityUsage.WithMessage("%s is not possible. No non-zero unique fields found in given model [%v]", "CheckUniqueness", v)
	case len(exprs) == 1:
		where = clause.Where{Exprs: []clause.Expression{exprs[0]}}
	default:
		where = clause.Where{Exprs: []clause.Expression{clause.Or(exprs...)}}
	}

	// fetch and parse result
	existing := reflect.New(resolver.ModelType()).Interface()
	switch rs := g.DB(ctx).Where(where).Take(existing); {
	case errors.Is(rs.Error, gorm.ErrRecordNotFound):
		return nil, nil
	case rs.Error != nil:
		return nil, rs.Error
	}

	// find duplicates
	for _, m := range models {
		dups = findDuplicateFields(m, existing, uniqueKeys)
		if len(dups) != 0 {
			break
		}
	}
	pairs := make([]string, 0, len(dups))
	for k, v := range dups {
		pairs = append(pairs, fmt.Sprintf("%s=[%v]", k, v))
	}
	return dups, data.ErrorDuplicateKey.WithMessage("entity with following properties already exists: %s", strings.Join(pairs, ", "))
}

/************************
	Helpers
 ************************/

func findUniqueIndexes(s *schema.Schema) []index {
	indexes := s.ParseIndexes()
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
	return uniqueness
}

func resolveIndexes(s *schema.Schema, keys []interface{}) ([]index, error) {
	ret := make([]index, 0, len(keys))
	for _, k := range keys {
		var idx index
		var e error
		switch v := k.(type) {
		case string:
			idx, e = asIndex(s, []string{v})
		case []string:
			idx, e = asIndex(s, v)
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

func asIndex(s *schema.Schema, names []string) (index, error) {
	ret := make(index, len(names))
	for i, n := range names {
		f, paths := lookupField(s, n)
		switch {
		case f == nil:
			return nil, fmt.Errorf("field with name [%s] is not found on model %s", n, s.Name)
		case len(paths) > 0:
			return nil, fmt.Errorf("associations are not supported in this utils")
		}
		ret[i] = &schema.IndexOption{Field: f}
	}
	return ret, nil
}

func uniquenessWhere(models []interface{}, keys []index) (exprs []clause.Expression) {
	for _, m := range models {
		exprs = append(exprs, uniquenessExprs(reflect.ValueOf(m), keys)...)
	}
	return
}

func uniquenessExprs(modelV reflect.Value, keys []index) []clause.Expression {
	sort.SliceStable(keys, func(i, j int) bool {
		return len(keys[i]) < len(keys[j])
	})
	exprs := make([]clause.Expression, 0, len(keys))
	modelV = reflect.Indirect(modelV)
	for _, idx := range keys {
		if expr, ok := compositeEqExpr(modelV, idx); ok {
			exprs = append(exprs, expr)
		}
	}
	return exprs
}

// compositeEqExpr returns false if
// 1. any index values are zero values
// 2. any column value is not found
// 3. len(cols) == 0
func compositeEqExpr(modelV reflect.Value, idx index) (clause.Expression, bool) {
	andExprs := make([]clause.Expression, len(idx))
	for i, f := range idx {
		v, ok := extractValue(modelV, f.Field)
		if !ok || !v.IsValid() || v.IsZero() {
			return nil, false
		}
		andExprs[i] = clause.Eq{
			Column: clause.Column{Name: f.DBName},
			Value:  v.Interface(),
		}
	}
	switch {
	case len(andExprs) == 0:
		return nil, false
	case len(andExprs) == 1:
		return andExprs[0], true
	default:
		return clause.And(andExprs...), true
	}
}

// findDuplicateFields compare fields and returns fields that left and right are same
func findDuplicateFields(left interface{}, right interface{}, keys []index) map[string]interface{} {
	dups := map[string]interface{}{}
	leftV := reflect.Indirect(reflect.ValueOf(left))
	rightV := reflect.Indirect(reflect.ValueOf(right))
	for _, idx := range keys {
		for _, f := range idx {
			lVal, lok := extractValue(leftV, f.Field)
			if !lok {
				continue
			}
			rVal, rok := extractValue(rightV, f.Field)
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

func extractValue(modelV reflect.Value, f *schema.Field) (reflect.Value, bool) {
	switch modelV.Kind() {
	case reflect.Map:
		for i := modelV.MapRange(); i.Next(); {
			k, ok := i.Key().Interface().(string)
			if ok && (k == f.Name || k == f.DBName) {
				return i.Value(), true
			}
		}
		return reflect.Value{}, false
	case reflect.Struct:
		return modelV.FieldByIndex(f.StructField.Index), true
	default:
		return reflect.Value{}, false
	}
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
	default:
		return nil
	}
}
