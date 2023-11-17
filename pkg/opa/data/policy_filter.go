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
	"sync"
)

/****************************
	Types
 ****************************/

// FilteredModel is a marker type that can be used in model struct as Embedded Struct.
// It's responsible for automatically applying OPA policy-based data filtering on model fields with "opa" tag.
//
// FilteredModel uses following GORM interfaces to modify PostgreSQL statements during select/update/delete,
// and apply value checks during create/update:
//
// 	- schema.QueryClausesInterface
// 	- schema.UpdateClausesInterface
// 	- schema.DeleteClausesInterface
// 	- schema.CreateClausesInterface
//
// When FilteredModel is present in data model, any model's fields tagged with "opa" will be used by OPA engine as following:
//
//	- During "create", values are included with path "input.resource.<opa_field_name>"
// 	- During "update", values are included with path "input.resources.delta.<opa_field_name>"
// 	- During "select/update/delete", "input.resources.<opa_field_name>" is used as "unknowns" during OPA Partial Evaluation,
// 	  and the result is translated to "WHERE" clause in PostgreSQL
//
// Where "<opa_field_name>" is specified by "opa" tag as `opa:"field:<opa_field_name>"`
//
// # Usage:
//
// FilteredModel is used as Embedded Struct in model struct,
//  - "opa" tag is required with resource type defined:
// 	  	`opa:"type:<opa_res_type>"`
//	- "gorm" tag should not be applied to the embedded struct
//
// # Examples:
//
// 	type Model struct {
//		ID              uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
//		Value           string
//		TenantID        uuid.UUID            `gorm:"type:KeyID;not null" opa:"field:tenant_id"`
//		TenantPath      pqx.UUIDArray        `gorm:"type:uuid[];index:,type:gin;not null" opa:"field:tenant_path"`
//		OwnerID         uuid.UUID            `gorm:"type:KeyID;not null" opa:"field:owner_id"`
//		opadata.FilteredModel `opa:"type:my_resource"`
// 	}
//
// Note: OPA filtering on relationships are currently not supported
//
// # Supported Tags:
//
// OPA tag should be in format of:
//	  `opa:"<key>:<value,<key>:<value>,..."`
// Invalid format or use of unsupported tag keys will result schema parsing error.
//
// Supported tag keys are:
// 	- "field:<opa_input_field_name>": required on any data field in model, only applicable on data fields
// 	- "input:<opa_input_field_name>": "input" is an alias of "field", only applicable on data fields
// 	- "type:<opa_resource_type>": required on FilteredModel. Ignored on other fields.
//    This value will be used as prefix/package of OPA policy: e.g. "<opa_resource_type>/<policy_name>"
//
// Following keys can override CRUD policies and only applicable on FilteredModel:
//
// 		+ "create:<policy_name>": optional, override policy used in OPA during create.
// 		+ "read:<policy_name>": optional, override policy used in OPA during read.
// 		+ "update:<policy_name>": optional, override policy used in OPA during update.
// 		+ "delete:<policy_name>": optional, override policy used in OPA during delete.
//  	+ "package:<policy_package>": optional, override policy's package. Default is "resource.<opa_resource_type>"
//
// Note: When <policy_name> is "-", policy-based data filtering is disabled for that operation.
// The default values are "filter_<op>"
type FilteredModel struct{
	PolicyFilter policyFilter `gorm:"-"`
}

// Filter is a marker type that can be used in model struct as Struct Field.
// It's responsible for automatically applying OPA policy-based data filtering on model fields with "opa" tag.
//
// Filter uses following GORM interfaces to modify PostgreSQL statements during select/update/delete,
// and apply value checks during create/update:
//
// 	- schema.QueryClausesInterface
// 	- schema.UpdateClausesInterface
// 	- schema.DeleteClausesInterface
// 	- schema.CreateClausesInterface
//
// When Filter is present in data model, any model's fields tagged with "opa" will be used by OPA engine as following:
//
//	- During "create", values are included with path "input.resource.<opa_field_name>"
// 	- During "update", values are included with path "input.resources.delta.<opa_field_name>"
// 	- During "select/update/delete", "input.resources.<opa_field_name>" is used as "unknowns" during OPA Partial Evaluation,
// 	  and the result is translated to "WHERE" clause in PostgreSQL
//
// Where "<opa_field_name>" is specified by "opa" tag as `opa:"field:<opa_field_name>"`
//
// # Usage:
//
// Filter is used as type of Struct Field within model struct:
//	- "opa" tag is required on the field with resource type defined:
// 	  	`opa:"type:<opa_res_type>"`
//	- `gorm:"-"` is required
//  - the field need to be exported
//
// # Examples:
//
// 	type Model struct {
//		ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
//		Value     string
//		OwnerName string
//		OwnerID   uuid.UUID            `gorm:"type:KeyID;not null" opa:"field:owner_id"`
//		Sharing   constraints.Sharing  `opa:"field:sharing"`
//		OPAFilter opadata.Filter `gorm:"-" opa:"type:my_resource"`
// 	}
//
// Note: OPA filtering on relationships are currently not supported
//
// # Supported Tags:
//
// OPA tag should be in format of:
//	  `opa:"<key>:<value,<key>:<value>,..."`
// Invalid format or use of unsupported tag keys will result schema parsing error.
//
// Supported tag keys are:
// 	- "field:<opa_input_field_name>": required on any data field in model, only applicable on data fields
// 	- "input:<opa_input_field_name>": "input" is an alias of "field", only applicable on data fields
// 	- "type:<opa_resource_type>": required on FilteredModel. Ignored on other fields.
//    This value will be used as prefix/package of OPA policy: e.g. "<opa_resource_type>/<policy_name>"
//
// Following keys can override CRUD policies and only applicable on FilteredModel:
//
// 		+ "create:<policy_name>": optional, override policy used in OPA during create.
// 		+ "read:<policy_name>": optional, override policy used in OPA during read.
// 		+ "update:<policy_name>": optional, override policy used in OPA during update.
// 		+ "delete:<policy_name>": optional, override policy used in OPA during delete.
//  	+ "package:<policy_package>": optional, override policy's package. Default is "resource.<opa_resource_type>"
//
// Note: When <policy_name> is "-", policy-based data filtering is disabled for that operation.
// The default values are "filter_<op>"
type Filter struct{
	policyFilter
}

/****************************
	Policy Filter
 ****************************/

// policyFilter implements
// - schema.GormDataTypeInterface
// - schema.QueryClausesInterface
// - schema.UpdateClausesInterface
// - schema.DeleteClausesInterface
// - schema.CreateClausesInterface
// this data type adds "WHERE" clause for OPA policy filtering
// Note: policyFilter should be used in model struct as a named field with `gorm:"-"` tag
type policyFilter struct{}

// QueryClauses implements schema.QueryClausesInterface,
func (pf policyFilter) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newStatementModifier(f, DBOperationFlagRead)}
}

// UpdateClauses implements schema.UpdateClausesInterface,
func (pf policyFilter) UpdateClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newStatementModifier(f, DBOperationFlagUpdate)}
}

// DeleteClauses implements schema.DeleteClausesInterface,
func (pf policyFilter) DeleteClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newStatementModifier(f, DBOperationFlagDelete)}
}

// CreateClauses implements schema.CreateClausesInterface,
func (pf policyFilter) CreateClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newStatementModifier(f, DBOperationFlagCreate)}
}

/***************************
	Read, Delete
 ***************************/

// statementModifier implements clause.Interface and gorm.StatementModifier, where gorm.StatementModifier do the real work.
// See gorm.DeletedAt for impl. reference
type statementModifier struct {
	types.NoopStatementModifier
	Metadata
	initOnce             sync.Once
	Schema               *schema.Schema
	Flag                 DBOperationFlag
	OPAFilterOptionsFunc func(stmt *gorm.Statement) (opa.ResourceFilterOptions, error)
}

func newStatementModifier(f *schema.Field, flag DBOperationFlag) clause.Interface {
	switch flag {
	case DBOperationFlagCreate:
		return newCreateStatementModifier(f.Schema)
	case DBOperationFlagUpdate:
		return newUpdateStatementModifier(f.Schema)
	default:
		ret := &statementModifier{
			Schema: f.Schema,
			Flag:   flag,
		}
		ret.OPAFilterOptionsFunc = ret.opaFilterOptions
		return ret
	}
}

func (m *statementModifier) lazyInit() (err error) {
	m.initOnce.Do(func() {
		if ptr, e := loadMetadata(m.Schema); e != nil {
			err = data.NewDataError(data.ErrorCodeInvalidApiUsage, e)
		} else {
			m.Metadata = *ptr
		}
	})
	return
}

func (m *statementModifier) ModifyStatement(stmt *gorm.Statement) {
	if stmt.AddError(m.lazyInit()) != nil {
		return
	}

	if shouldSkip(stmt.Context, m.Flag, m.mode) {
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
		case errors.Is(e, opa.ErrQueriesNotResolved):
			_ = stmt.AddError(opa.ErrAccessDenied.WithMessage("record not found"))
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

func (m *statementModifier) opaFilterOptions(stmt *gorm.Statement) (opa.ResourceFilterOptions, error) {
	unknowns := make([]string, 0, len(m.Fields))
	for k := range m.Fields {
		unknown := fmt.Sprintf(`%s.%s.%s`, opa.InputPrefixRoot, opa.InputPrefixResource, k)
		unknowns = append(unknowns, unknown)
	}
	return func(rf *opa.ResourceFilter) {
		rf.Query = resolveQuery(stmt.Context, m.Flag, true, &m.Metadata)
		rf.Unknowns = unknowns
		rf.QueryMapper = NewGormPartialQueryMapper(&GormMapperConfig{
			Metadata:  &m.Metadata,
			Fields:    m.Fields,
			Statement: stmt,
		})
		populateExtraData(stmt.Context, rf.ExtraData)
	}, nil
}

/***************************
	Update
 ***************************/

// updateStatementModifier is a special statementModifier that TODO
type updateStatementModifier struct {
	statementModifier
}

func newUpdateStatementModifier(s *schema.Schema) *updateStatementModifier {
	ret := &updateStatementModifier{
		statementModifier{
			Schema: s,
			Flag:   DBOperationFlagUpdate,
		},
	}
	ret.OPAFilterOptionsFunc = ret.opaFilterOptions
	return ret
}

func (m *updateStatementModifier) opaFilterOptions(stmt *gorm.Statement) (opa.ResourceFilterOptions, error) {
	opts, e := m.statementModifier.opaFilterOptions(stmt)
	if e != nil {
		return nil, e
	}
	models, e := resolvePolicyTargets(stmt, &m.Metadata, m.Flag)
	if e != nil {
		return nil, ErrUnsupportedUsage.WithMessage("failed resolve delta in 'update' DB operation: %v", e)
	}
	switch len(models) {
	case 1:
		break
	case 0:
		return nil, ErrUnsupportedUsage.WithMessage("unable to resolve delta in 'update' DB operation")
	default:
		return nil, ErrUnsupportedUsage.WithMessage("'update' DB operation with batch update is not supported")
	}
	values, e := models[0].toResourceValues()
	if e != nil {
		return opts, ErrUnsupportedUsage.WithMessage(`%v`, e)
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

func newCreateStatementModifier(s *schema.Schema) *createStatementModifier {
	return &createStatementModifier{
		statementModifier{
			Schema: s,
			Flag:   DBOperationFlagCreate,
		},
	}
}

func (m *createStatementModifier) ModifyStatement(stmt *gorm.Statement) {
	if stmt.AddError(m.lazyInit()) != nil {
		return
	}

	if shouldSkip(stmt.Context, DBOperationFlagCreate, m.mode) {
		return
	}

	models, e := resolvePolicyTargets(stmt, &m.Metadata, m.Flag)
	if stmt.Statement.AddError(e) != nil {
		return
	}
	for i := range models {
		if stmt.Statement.AddError(m.checkPolicy(stmt.Context, &models[i])) != nil {
			return
		}
	}
}

func (m *createStatementModifier) checkPolicy(ctx context.Context, model *policyTarget) error {
	values, e := model.toResourceValues()
	if e != nil {
		return opa.ErrAccessDenied.WithMessage(`Cannot resolve values for model creation`)
	}
	return opa.AllowResource(ctx, model.meta.ResType, opa.OpCreate, func(res *opa.ResourceQuery) {
		res.ResourceValues = *values
		res.Policy = resolveQuery(ctx, m.Flag, false, &m.Metadata)
		populateExtraData(ctx, res.ExtraData)
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
