package types

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"database/sql/driver"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// TenantPath implements
// - schema.GormDataTypeInterface
// - schema.QueryClausesInterface
// - schema.UpdateClausesInterface
// - schema.DeleteClausesInterface
// this data type adds "WHERE" clause for tenancy filtering
type TenantPath pqx.UUIDArray

// Value implements driver.Valuer
func (t TenantPath) Value() (driver.Value, error) {
	return pqx.UUIDArray(t).Value()
}

// Scan implements sql.Scanner
func (t *TenantPath) Scan(src interface{}) error {
	return (*pqx.UUIDArray)(t).Scan(src)
}

func (t TenantPath) GormDataType() string {
	return "uuid[]"
}

// QueryClauses implements schema.QueryClausesInterface,
func (t TenantPath) QueryClauses(_ *schema.Field) []clause.Interface {
	// TODO return tenancyFilterClause if we want this for SELECT statement
	return []clause.Interface{}
}

// UpdateClauses implements schema.UpdateClausesInterface,
func (t TenantPath) UpdateClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{&tenancyFilterClause{Field: f}}
}

// DeleteClauses implements schema.DeleteClausesInterface,
func (t TenantPath) DeleteClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{&tenancyFilterClause{Field: f}}
}

// tenancyFilterClause implements clause.Interface and gorm.StatementModifier, where gorm.StatementModifier do the real work.
// See gorm.DeletedAt for impl. reference
type tenancyFilterClause struct {
	stmtModifier
	Field *schema.Field
}

func (c tenancyFilterClause) ModifyStatement(stmt *gorm.Statement) {
	if shouldSkipTenancyCheck(stmt.Context) {
		return
	}

	tenantIDs := requiredTenancyFiltering(stmt)
	if len(tenantIDs) == 0 {
		return
	}

	// special fix for db.Model(&model{}).Where(&model{f1:v1}).Or(&model{f2:v2})...
	// Ref:	https://github.com/go-gorm/gorm/issues/3627
	//		https://github.com/go-gorm/gorm/commit/9b2181199d88ed6f74650d73fa9d20264dd134c0#diff-e3e9193af67f3a706b3fe042a9f121d3609721da110f6a585cdb1d1660fd5a3c
	fixWhereClausesForStatementModifier(stmt)

	// add tenancy filter condition
	colExpr := stmt.Quote(clause.Column{ Table: clause.CurrentTable, Name:  c.Field.DBName })
	sql := fmt.Sprintf("%s @> ?", colExpr)
	var conditions []clause.Expression
	for _, id := range tenantIDs {
		conditions = append(conditions, clause.Expr{
			SQL:  sql,
			Vars: []interface{}{pqx.UUIDArray{id}},
		})
	}
	if len(conditions) == 1 {
		stmt.AddClause(clause.Where{Exprs: conditions})
	} else {
		stmt.AddClause(clause.Where{Exprs: []clause.Expression{clause.Or(conditions...)}})
	}
}

/***********************
	Helpers
 ***********************/

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
