package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

/****************************
	Func
 ****************************/

/****************************
	Types
 ****************************/

// PolicyFilter implements
// - schema.GormDataTypeInterface
// - schema.QueryClausesInterface
// - schema.UpdateClausesInterface
// - schema.DeleteClausesInterface
// - schema.CreateClausesInterface
// this data type adds "WHERE" clause for tenancy filtering
type PolicyFilter struct{}

// QueryClauses implements schema.QueryClausesInterface,
func (pf PolicyFilter) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newStatementModifier(f, DBOperationFlagRead)}
}

// UpdateClauses implements schema.UpdateClausesInterface,
func (pf PolicyFilter) UpdateClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newStatementModifier(f, DBOperationFlagUpdate)}
}

// DeleteClauses implements schema.DeleteClausesInterface,
func (pf PolicyFilter) DeleteClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newStatementModifier(f, DBOperationFlagDelete)}
}

// CreateClauses implements schema.CreateClausesInterface,
func (pf PolicyFilter) CreateClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newStatementModifier(f, DBOperationFlagCreate)}
}

/***************************
	Read, Update, Delete
 ***************************/

// statementModifier implements clause.Interface and gorm.StatementModifier, where gorm.StatementModifier do the real work.
// See gorm.DeletedAt for impl. reference
type statementModifier struct {
	types.NoopStatementModifier
	metadata
	Flag                 DBOperationFlag
	OPAFilterOptionsFunc func(stmt *gorm.Statement) (opa.ResourceFilterOptions, error)
}

func newStatementModifier(f *schema.Field, flag DBOperationFlag) clause.Interface {
	meta, e := loadMetadata(f.Schema)
	if e != nil {
		panic(e)
	}
	switch flag {
	case DBOperationFlagCreate:
		return newCreateStatementModifier(meta)
	case DBOperationFlagUpdate:
		return newUpdateStatementModifier(meta)
	default:
		ret := &statementModifier{
			metadata: *meta,
			Flag:     flag,
		}
		ret.OPAFilterOptionsFunc = ret.opaFilterOptions
		return ret
	}
}

func (m statementModifier) ModifyStatement(stmt *gorm.Statement) {
	if shouldSkip(stmt.Context, m.Flag, m.Mode) {
		return
	}

	filterOpts, e := m.OPAFilterOptionsFunc(stmt)
	if e != nil {
		_ = stmt.AddError(data.NewDataError(data.ErrorCodeInvalidApiUsage, fmt.Sprintf(`OPA filtering failed with error: %v`, e), e))
		return
	}
	rs, e := opa.FilterResource(stmt.Context, m.ResType, flagToResOp(m.Flag), filterOpts)
	if e != nil {
		switch {
		case errors.Is(e, opa.QueriesNotResolvedError):
			_ = stmt.AddError(data.NewRecordNotFoundError("record not found"))
		default:
			_ = stmt.AddError(data.NewInternalError(fmt.Sprintf(`OPA filtering failed with error: %v`, e), e))
		}
		return
	}
	exprs := rs.Result.([]clause.Expression)
	if len(exprs) == 0 {
		return
	}

	// special fix for db.Model(&policyTarget{}).Where(&policyTarget{f1:v1}).Or(&policyTarget{f2:v2})...
	// Ref:	https://github.com/go-gorm/gorm/issues/3627
	//		https://github.com/go-gorm/gorm/commit/9b2181199d88ed6f74650d73fa9d20264dd134c0#diff-e3e9193af67f3a706b3fe042a9f121d3609721da110f6a585cdb1d1660fd5a3c
	types.FixWhereClausesForStatementModifier(stmt)

	if len(exprs) == 1 {
		stmt.AddClause(clause.Where{Exprs: exprs})
	} else {
		stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.Or(exprs...)}})
	}
}

func (m statementModifier) opaFilterOptions(stmt *gorm.Statement) (opa.ResourceFilterOptions, error) {
	unknowns := make([]string, 0, len(m.Fields))
	for k := range m.Fields {
		unknown := fmt.Sprintf(`%s.%s.%s`, opa.InputPrefixRoot, opa.InputPrefixResource, k)
		unknowns = append(unknowns, unknown)
	}
	return func(rf *opa.ResourceFilter) {
		rf.Policy = m.Policy
		rf.Unknowns = unknowns
		rf.QueryMapper = NewGormPartialQueryMapper(&GormMapperConfig{
			Fields:    m.Fields,
			Statement: stmt,
		})
	}, nil
}

/***************************
	Update
 ***************************/

// updateStatementModifier is a special statementModifier that TODO
type updateStatementModifier struct {
	statementModifier
}

func newUpdateStatementModifier(meta *metadata) *updateStatementModifier {
	ret := &updateStatementModifier{
		statementModifier{
			metadata: *meta,
			Flag:     DBOperationFlagUpdate,
		},
	}
	ret.OPAFilterOptionsFunc = ret.opaFilterOptions
	return ret
}

func (m updateStatementModifier) opaFilterOptions(stmt *gorm.Statement) (opa.ResourceFilterOptions, error) {
	opts, e := m.statementModifier.opaFilterOptions(stmt)
	if e != nil {
		return nil, e
	}
	models, e := resolvePolicyTargets(stmt, &m.metadata, m.Flag)
	if e != nil {
		return nil, UnsupportedUsageError.WithMessage("failed resolve delta in 'update' DB operation: %v", e)
	}
	switch len(models) {
	case 1:
		break
	case 0:
		return nil, UnsupportedUsageError.WithMessage("unable to resolve delta in 'update' DB operation")
	default:
		return nil, UnsupportedUsageError.WithMessage("'update' DB operation with batch update is not supported")
	}
	values, e := models[0].toResourceValues()
	if e != nil {
		return opts, UnsupportedUsageError.WithMessage(`%v`, e)
	}
	return func(rf *opa.ResourceFilter) {
		opts(rf)
		rf.Delta = values
	}, nil
}

/***************************
	Create
 ***************************/

// createStatementModifier is a special statementModifier that perform OPA policy check on resource creation
// Note: this modifier doesn't actually modify statement, it checks the to-be-created model/map against OPA and
// 		 returns error if not allowed
type createStatementModifier struct {
	statementModifier
}

func newCreateStatementModifier(meta *metadata) *createStatementModifier {
	return &createStatementModifier{
		statementModifier{
			metadata: *meta,
			Flag:     DBOperationFlagCreate,
		},
	}
}

func (m createStatementModifier) ModifyStatement(stmt *gorm.Statement) {
	if shouldSkip(stmt.Context, DBOperationFlagCreate, m.Mode) {
		return
	}

	models, e := resolvePolicyTargets(stmt, &m.metadata, m.Flag)
	if stmt.Statement.AddError(e) != nil {
		return
	}
	for _, model := range models {
		if stmt.Statement.AddError(m.checkPolicy(stmt.Context, &model)) != nil {
			return
		}
	}
}

func (m createStatementModifier) checkPolicy(ctx context.Context, model *policyTarget) error {
	values, e := model.toResourceValues()
	if e != nil {
		return opa.AccessDeniedError.WithMessage(`Cannot resolve values for model creation`)
	}
	return opa.AllowResource(ctx, model.meta.ResType, opa.OpCreate, func(res *opa.Resource) {
		res.ResourceValues = *values
	})
}

/***********************
	Helpers
 ***********************/

func flagToResOp(flag DBOperationFlag) opa.ResourceOperation {
	switch flag {
	case DBOperationFlagRead:
		return opa.OpRead
	case DBOperationFlagUpdate:
		return opa.OpWrite
	case DBOperationFlagDelete:
		return opa.OpDelete
	default:
		return opa.OpCreate
	}
}
