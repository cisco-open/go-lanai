package opadata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/reflectutils"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
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
// this data type adds "WHERE" clause for tenancy filtering
type PolicyFilter struct {

}

// QueryClauses implements schema.QueryClausesInterface,
func (pf PolicyFilter) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newPolicyFilterClause(f, FilteringFlagReadFiltering)}
}

// UpdateClauses implements schema.UpdateClausesInterface,
func (pf PolicyFilter) UpdateClauses(f *schema.Field) []clause.Interface {
	//TODO
	//return []clause.Interface{newPolicyFilterClause(f, FilteringFlagWriteFiltering)}
	return []clause.Interface{}
}

// DeleteClauses implements schema.DeleteClausesInterface,
func (pf PolicyFilter) DeleteClauses(f *schema.Field) []clause.Interface {
	// TODO
	//return []clause.Interface{newPolicyFilterClause(f, FilteringFlagDeleteFiltering)}
	return []clause.Interface{}
}

// policyFilterClause implements clause.Interface and gorm.StatementModifier, where gorm.StatementModifier do the real work.
// See gorm.DeletedAt for impl. reference
type policyFilterClause struct {
	types.NoopStatementModifier
	Flag   FilteringFlag
	Mode   filteringMode
	Fields map[string]*schema.Field
	Schema *schema.Schema
}

func newPolicyFilterClause(f *schema.Field, flag FilteringFlag) *policyFilterClause {
	// TODO determine mode
	return &policyFilterClause{
		Flag:   flag,
		Mode:   determineFilteringMode(f),
		Fields: collectFields(f.Schema),
		Schema: f.Schema,
	}
}

func (c policyFilterClause) ModifyStatement(stmt *gorm.Statement) {
	if shouldSkip(stmt.Context, c.Flag, c.Mode) {
		return
	}

	rs, e := opa.FilterResource(stmt.Context, "poc", flagToResOp(c.Flag), c.opaFilterOptions(stmt))
	if e != nil {
		switch {
		case errors.Is(e, opa.QueriesNotResolvedError):
			stmt.Error = data.NewRecordNotFoundError("record not found")
		default:
			stmt.Error = data.NewInternalError(fmt.Sprintf(`OPA filtering failed with error: %v`, e), e)
		}
		return
	}
	exprs := rs.Result.([]clause.Expression)
	if len(exprs) == 0 {
		return
	}

	// special fix for db.Model(&model{}).Where(&model{f1:v1}).Or(&model{f2:v2})...
	// Ref:	https://github.com/go-gorm/gorm/issues/3627
	//		https://github.com/go-gorm/gorm/commit/9b2181199d88ed6f74650d73fa9d20264dd134c0#diff-e3e9193af67f3a706b3fe042a9f121d3609721da110f6a585cdb1d1660fd5a3c
	types.FixWhereClausesForStatementModifier(stmt)

	if len(exprs) == 1 {
		stmt.AddClause(clause.Where{Exprs: exprs})
	} else {
		stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.Or(exprs...)}})
	}
}

func (c policyFilterClause) opaFilterOptions(stmt *gorm.Statement) opa.ResourceFilterOptions {
	unknowns := make([]string, 0, len(c.Fields))
	for k := range c.Fields {
		unknown := fmt.Sprintf(`%s.%s.%s`, opa.InputPrefixRoot, opa.InputPrefixResource, k)
		unknowns = append(unknowns, unknown)
	}
	return func(rf *opa.ResourceFilter) {
		rf.Unknowns = unknowns
		rf.QueryMapper = NewGormPartialQueryMapper(&GormMapperConfig{
			Fields:    c.Fields,
			Statement: stmt,
		})
	}
}

/***********************
	Helpers
 ***********************/

func extractFilterTag(f *schema.Field) string {
	if tag, ok := f.Tag.Lookup(types.TagFilter); ok {
		return strings.ToLower(strings.TrimSpace(tag))
	}
	// TODO Fix this: check if tag is available on embedded struct
	sf, ok := reflectutils.FindStructField(f.Schema.ModelType, func(t reflect.StructField) bool {
		return t.Anonymous && (t.Type.AssignableTo(typeTenancy) || t.Type.AssignableTo(typeTenancyPtr))
	})
	if ok {
		return sf.Tag.Get(types.TagFilter)
	}
	return ""
}

func determineFilteringMode(f *schema.Field) (mode filteringMode) {
	// TODO determine mode
	mode = filteringMode(FilteringFlagWriteValueCheck)
	tag := extractFilterTag(f)
	switch tag {
	case "":
		mode = filteringModeDefault
	case "-":
	default:
		if strings.ContainsRune(tag, 'r') {
			mode = mode | filteringMode(FilteringFlagReadFiltering)
		}
		if strings.ContainsRune(tag, 'w') {
			mode = mode | filteringMode(FilteringFlagWriteFiltering)
		}
	}
	return
}

func collectFields(s *schema.Schema) (ret map[string]*schema.Field) {
	ret = map[string]*schema.Field{}
	for _, f := range s.Fields {
		if tag, ok := f.Tag.Lookup(TagOPA); ok {
			ret[strings.TrimSpace(tag)] = f
		}
	}
	return
}

func flagToResOp(flag FilteringFlag) opa.ResourceOperation {
	switch flag {
	case FilteringFlagReadFiltering:
		return opa.OpRead
	case FilteringFlagWriteFiltering:
		return opa.OpWrite
	case FilteringFlagDeleteFiltering:
		return opa.OpDelete
	default:
		return opa.OpCreate
	}
}

func requiredTenancyFiltering(stmt *gorm.Statement) (tenantIDs []uuid.UUID) {
	auth := security.Get(stmt.Context)
	if security.HasPermissions(auth, security.SpecialPermissionAccessAllTenant) {
		return nil
	}

	ud, _ := auth.Details().(security.UserDetails)
	if ud != nil {
		idsStr := ud.AssignedTenantIds()
		tenantIDs = make([]uuid.UUID, 0, len(idsStr))
		for tenant := range idsStr {
			if tenantId, e := uuid.Parse(tenant); e == nil {
				tenantIDs = append(tenantIDs, tenantId)
			}
		}
	}
	return
}
