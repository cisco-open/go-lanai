package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/regoexpr"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/sdk"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

var (
	typeUUID          = reflect.TypeOf(uuid.Nil)
	typeUUIDPtr       = reflect.TypeOf(&uuid.UUID{})
	typeUUIDArray     = reflect.TypeOf(pqx.UUIDArray{})

	typeTenantPath    = reflect.TypeOf(PolicyFilter{})
	typeTenancy       = reflect.TypeOf(Tenancy{})
	typeTenancyPtr    = reflect.TypeOf(&Tenancy{})
	mapKeysTenantID   = utils.NewStringSet(fieldTenantID, colTenantID)
	mapKeysTenantPath = utils.NewStringSet(fieldTenantPath, colTenantPath)
)

type GormMapperConfig struct {
	Fields    map[string]*schema.Field
	Statement *gorm.Statement
}

type GormPartialQueryMapper struct {
	ctx    context.Context
	fields map[string]*schema.Field
	stmt   *gorm.Statement
}

func NewGormPartialQueryMapper(cfg *GormMapperConfig) *GormPartialQueryMapper {
	return &GormPartialQueryMapper{
		ctx:    context.Background(),
		fields: cfg.Fields,
		stmt:   cfg.Statement,
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
	f, e := m.Field(ctx, colRef)
	if e != nil {
		return nil, e
	}
	col := clause.Column{
		Table: f.Schema.Table,
		Name:  f.DBName,
	}
	if val, e = m.Value(ctx, f, val); e != nil {
		return nil, e
	}

	switch op.Hash() {
	case regoexpr.OpHashEqual, regoexpr.OpHashEq:
		ret = &clause.Eq{Column: col, Value: val}
	case regoexpr.OpHashNeq:
		ret = &clause.Neq{Column: col, Value: val}
	case regoexpr.OpHashLte:
		ret = &clause.Lte{Column: col, Value: val}
	case regoexpr.OpHashLt:
		ret = &clause.Lt{Column: col, Value: val}
	case regoexpr.OpHashGte:
		ret = &clause.Gte{Column: col, Value: val}
	case regoexpr.OpHashGt:
		ret = &clause.Gt{Column: col, Value: val}
	case regoexpr.OpHashIn:
		// TODO should use statement quote
		sql := fmt.Sprintf("%s @> ?", m.Quote(ctx, col))
		ret = clause.Expr{
			SQL:  sql,
			Vars: []interface{}{val},
		}
	default:
		return nil, QueryTranslationError.WithMessage("Unsupported Rego operator: %v", op)
	}
	return
}

/****************
	Helpers
 ****************/

func (m *GormPartialQueryMapper) Quote(_ context.Context, field interface{}) string {
	return m.stmt.Quote(field)
}

func (m *GormPartialQueryMapper) Field(_ context.Context, colRef ast.Ref) (ret *schema.Field, err error) {
	// TODO review this part
	path := colRef.String()
	idx := strings.LastIndex(path, ".")
	f, ok := m.fields[path[idx+1:]]
	if !ok {
		return ret, QueryTranslationError.WithMessage(`unable to resolve column with OPA unknowns [%s]`, path)
	}
	return f, nil
}

func (m *GormPartialQueryMapper) Value(_ context.Context, f *schema.Field, val interface{}) (interface{}, error) {
	if rv := reflect.ValueOf(val); rv.CanConvert(f.FieldType) {
		return rv.Convert(f.FieldType).Interface(), nil
	}
	switch ft := f.FieldType; {
	case ft == typeUUID:
		return m.toUUID(val)
	case ft == typeUUIDPtr:
		parsed, e := m.toUUID(val)
		if e != nil {
			return nil, e
		}
		return &parsed, nil
	case ft == typeUUIDArray:
		return m.toUUIDArray(val)
	default:
		return val, nil
	}
}

func (m *GormPartialQueryMapper) toUUID(val interface{}) (uuid.UUID, error) {
	switch v := val.(type) {
	case string:
		parsed, e := uuid.Parse(v)
		if e != nil {
			return uuid.Nil, e
		}
		return parsed, nil
	case uuid.UUID:
		return v, nil
	case *uuid.UUID:
		if v != nil {
			return *v, nil
		}
		return uuid.Nil, nil
	}
	return uuid.Nil, QueryTranslationError.WithMessage(`unable to convert [%v] to UUID`, val)
}

func (m *GormPartialQueryMapper) toUUIDArray(val interface{}) (pqx.UUIDArray, error) {
	switch v := val.(type) {
	case []string:
		uuids := pqx.UUIDArray(make([]uuid.UUID, len(v)))
		for i := range v {
			parsed, e := uuid.Parse(v[i])
			if e != nil {
				return nil, e
			}
			uuids[i] = parsed
		}
	case []uuid.UUID:
		return v, nil
	case []*uuid.UUID:
		uuids := pqx.UUIDArray(make([]uuid.UUID, 0, len(v)))
		for _, ptr := range v {
			if ptr != nil {
				uuids = append(uuids, *ptr)
			}
		}
	case string:
		parsed, e := uuid.Parse(v)
		if e != nil {
			return nil, e
		}
		return pqx.UUIDArray{parsed}, nil
	case uuid.UUID:
		return pqx.UUIDArray{v}, nil
	case *uuid.UUID:
		if v != nil {
			return pqx.UUIDArray{*v}, nil
		}
	}
	if val == nil {
		return pqx.UUIDArray{}, nil
	}
	return nil, QueryTranslationError.WithMessage(`unable to convert [%v] to UUID array`, val)
}
