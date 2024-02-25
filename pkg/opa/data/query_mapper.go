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

package opadata

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/opa/regoexpr"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/open-policy-agent/opa/ast"
    "github.com/open-policy-agent/opa/rego"
    "github.com/open-policy-agent/opa/sdk"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
    "gorm.io/gorm/schema"
    "reflect"
    "strings"
    "time"
)

var (
	typeScanner = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	colRefPrefix = ast.Ref{ast.VarTerm("input"), ast.StringTerm("resource")}
)

const (
	dataTypeJSONB = "jsonb"
)

type GormMapperConfig struct {
	Metadata  *Metadata
	Fields    map[string]*TaggedField
	Statement *gorm.Statement
}

type GormPartialQueryMapper struct {
	ctx       context.Context
	metadata  *Metadata
	fields    map[string]*TaggedField
	stmt      *gorm.Statement
}

func NewGormPartialQueryMapper(cfg *GormMapperConfig) *GormPartialQueryMapper {
	return &GormPartialQueryMapper{
		ctx:       context.Background(),
		metadata:  cfg.Metadata,
		fields:    cfg.Fields,
		stmt:      cfg.Statement,
	}
}

/*****************************
	ContextAware
 *****************************/

func (m *GormPartialQueryMapper) WithContext(ctx context.Context) sdk.PartialQueryMapper {
	mapper := *m
	mapper.ctx = ctx
	return &mapper
}

func (m *GormPartialQueryMapper) Context() context.Context {
	return m.ctx
}

/*****************************
	sdk.PartialQueryMapper
 *****************************/

func (m *GormPartialQueryMapper) MapResults(pq *rego.PartialQueries) (interface{}, error) {
	return regoexpr.TranslatePartialQueries(m.ctx, pq, func(opts *regoexpr.TranslateOption[clause.Expression]) {
		opts.Translator = m
	})
}

func (m *GormPartialQueryMapper) ResultToJSON(result interface{}) (interface{}, error) {
	data, e := json.Marshal(result)
	return string(data), e
}

/****************
	Translator
 ****************/

func (m *GormPartialQueryMapper) Negate(_ context.Context, expr clause.Expression) clause.Expression {
	return clause.Not(expr)
}

func (m *GormPartialQueryMapper) And(_ context.Context, exprs ...clause.Expression) clause.Expression {
	return clause.And(exprs...)
}

func (m *GormPartialQueryMapper) Or(_ context.Context, exprs ...clause.Expression) clause.Expression {
	return clause.Or(exprs...)
}

func (m *GormPartialQueryMapper) Comparison(ctx context.Context, op ast.Ref, colRef ast.Ref, val interface{}) (ret clause.Expression, err error) {
	field, path, e := m.ResolveField(ctx, colRef)
	if e != nil {
		return nil, e
	}
	colExpr := m.ResolveColumnExpr(ctx, field, path...)
	val = m.ResolveValueExpr(ctx, val, field)

	switch op.Hash() {
	case regoexpr.OpHashEqual, regoexpr.OpHashEq:
		ret = &clause.Eq{Column: colExpr, Value: val}
	case regoexpr.OpHashNeq:
		ret = &clause.Neq{Column: colExpr, Value: val}
	case regoexpr.OpHashLte:
		ret = &clause.Lte{Column: colExpr, Value: val}
	case regoexpr.OpHashLt:
		ret = &clause.Lt{Column: colExpr, Value: val}
	case regoexpr.OpHashGte:
		ret = &clause.Gte{Column: colExpr, Value: val}
	case regoexpr.OpHashGt:
		ret = &clause.Gt{Column: colExpr, Value: val}
	case regoexpr.OpHashIn:
		expr := fmt.Sprintf("%s @> ?", colExpr)
		ret = clause.Expr{
			SQL:  expr,
			Vars: []interface{}{val},
		}
	default:
		return nil, ErrQueryTranslation.WithMessage("Unsupported Rego operator: %v", op)
	}
	return
}

/****************
	Helpers
 ****************/

func (m *GormPartialQueryMapper) Quote(field interface{}) string {
	return m.stmt.Quote(field)
}

func (m *GormPartialQueryMapper) ResolveField(_ context.Context, colRef ast.Ref) (ret *TaggedField, jsonbPath []string, err error) {
	// TODO review this part
	if !colRef.HasPrefix(colRefPrefix) {
		return nil, nil, ErrQueryTranslation.WithMessage(`OPA unknowns [%v] is missing prefix "%v"`, colRef, colRefPrefix)
	}

	var field *TaggedField
	var key string
	var remaining []string
	for _, term := range colRef[len(colRefPrefix):] {
		var str string
		if e := ast.As(term.Value, &str); e != nil {
			return nil, nil, ErrQueryTranslation.WithMessage(`OPA unknowns [%v] contains invalid term [%v]`, colRef, term)
		}

		if field == nil {
			key = key + "." + str
			if key[0] == '.' {
				key = key[1:]
			}
			field, _ = m.fields[key]
		} else {
			remaining = append(remaining, str)
		}
	}
	if field == nil {
		return nil, nil, ErrQueryTranslation.WithMessage(`unable to resolve column with OPA unknowns [%v]`, colRef)
	}
	if len(remaining) != 0 && strings.ToLower(string(field.DataType)) != dataTypeJSONB {
		return nil, nil, ErrQueryTranslation.WithMessage(`unable to resolve column with OPA unknowns [%v]: found field "%s" but it's not JSONB`, colRef, field.Name)
	}
	return field, remaining, nil
}

// ResolveColumnExpr resolve column clause with given field and optional JSONB path
func (m *GormPartialQueryMapper) ResolveColumnExpr(_ context.Context, field *TaggedField, paths...string) string {
	col := clause.Column{
		Table: field.Schema.Table,
		Name:  field.DBName,
	}
	if len(paths) == 0 {
		return m.Quote(col)
	}
	// with remaining paths, the field is JSONB
	expr := m.Quote(col)
	for _, path := range paths {
		expr = fmt.Sprintf(`%s -> '%s'`, expr, path)
	}
	return expr
}

func (m *GormPartialQueryMapper) ResolveValueExpr(_ context.Context, val interface{}, field *TaggedField) interface{} {
	rv := reflect.ValueOf(val)
	// try convert using field's type
	if v, ok := m.resolveValueByType(rv, field.FieldType); ok {
		return v.Interface()
	}
	// fallback to presenting value to DB recognizable pattern based on data type
	if v, ok := m.resolveValueByDataType(rv, field.DataType); ok {
		return v.Interface()
	}
	return val
}

// resolveValueByType convert given src to DB recognizable value of given type hint. e.g. []string to pqx.UUIDArray.
// In case the type hint is potential reference or container of source value, the source value is converted and wrapped
// using type hint.
// 		e.g. pqx.UUIDArray is a potential container of string, the string is converted to a pqx.UUIDArray with single element
// 		e.g. uuid.UUID is a potential reference of string, the string is converted to a pointer to uuid.UUID
// Note: This function guarantees that the returned value is same type of given type hint
func (m *GormPartialQueryMapper) resolveValueByType(src reflect.Value, typeHint reflect.Type) (reflect.Value, bool) {
	// first, try convert directly, or via sql.Scanner API
	if resolved, ok := m.toType(src, typeHint); ok {
		return resolved, true
	}

	// second, if it's slice, array or pointer, try to convert given value to its Elem()
	var resolved reflect.Value
	kind := typeHint.Kind()
	//nolint:exhaustive // we only handle slice, array or pointer
	switch kind {
	case reflect.Slice, reflect.Array, reflect.Pointer:
		v, ok := m.resolveValueByType(src, typeHint.Elem())
		if !ok {
			return resolved, false
		}
		resolved = v
		// wrap resolved value into proper type
		//nolint:exhaustive // we only handle slice, array or pointer
		switch kind {
		case reflect.Slice:
			ret := reflect.MakeSlice(typeHint, 1, 1)
			ret.Index(0).Set(resolved)
			return ret, true
		case reflect.Pointer:
			if resolved.CanAddr() {
				return resolved.Addr(), true
			}
			return src, false
		case reflect.Array:
			ret := reflect.New(typeHint).Elem()
			if typeHint.Len() > 0 {
				ret.Index(0).Set(resolved)
			}
			return ret, true
		}
	}
	return src, false
}

// resolveValueByDataType try to present value to DB recognizable pattern based on data type.
// Note: we only support minimum set of data types.
func (m *GormPartialQueryMapper) resolveValueByDataType(src reflect.Value, dataType schema.DataType) (reflect.Value, bool) {
	switch strings.ToLower(string(dataType)) {
	case "jsonb":
		if data, e := json.Marshal(src.Interface()); e == nil {
			return reflect.ValueOf(string(data)), true
		}
	case string(schema.Time):
		if intValue, ok := m.toType(src, reflect.TypeOf(int64(0))); ok {
			// treat value as timestamp in seconds
			t := time.Unix(intValue.Int(), 0)
			return reflect.ValueOf(t), true
		}

		if src.Kind() == reflect.String {
			if t := utils.ParseTimeISO8601(src.String()); !t.IsZero() {
				return reflect.ValueOf(t), true
			}
		}
	}
	return src, false
}

// toType convert source value to given type using direct convert if it's scalar, string, alias, etc.,
// or using sql.Scanner interface
func (m *GormPartialQueryMapper) toType(src reflect.Value, typ reflect.Type) (reflect.Value, bool) {
	switch {
	case src.CanConvert(typ):
		return src.Convert(typ), true
	case typ.Implements(typeScanner):
		v := reflect.New(typ).Elem()
		if e := v.Interface().(sql.Scanner).Scan(src.Interface()); e == nil {
			return v, true
		}
	case reflect.PointerTo(typ).Implements(typeScanner):
		v := reflect.New(typ)
		if e := v.Interface().(sql.Scanner).Scan(src.Interface()); e == nil {
			return v.Elem(), true
		}
	}
	return src, false
}

