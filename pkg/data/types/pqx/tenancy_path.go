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

package pqx

import (
	"database/sql/driver"
	"fmt"
	securityinternal "github.com/cisco-open/go-lanai/internal/security"
	"github.com/cisco-open/go-lanai/pkg/data/types"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/utils/reflectutils"
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

// TenantPath implements
// - schema.GormDataTypeInterface
// - schema.QueryClausesInterface
// - schema.UpdateClausesInterface
// - schema.DeleteClausesInterface
// this data type adds "WHERE" clause for tenancy filtering
type TenantPath UUIDArray

// Value implements driver.Valuer
func (t TenantPath) Value() (driver.Value, error) {
	return UUIDArray(t).Value()
}

// Scan implements sql.Scanner
func (t *TenantPath) Scan(src interface{}) error {
	return (*UUIDArray)(t).Scan(src)
}

func (t TenantPath) GormDataType() string {
	return "uuid[]"
}

// QueryClauses implements schema.QueryClausesInterface,
func (t TenantPath) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newTenancyFilterClause(f, true)}
}

// UpdateClauses implements schema.UpdateClausesInterface,
func (t TenantPath) UpdateClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newTenancyFilterClause(f, false)}
}

// DeleteClauses implements schema.DeleteClausesInterface,
func (t TenantPath) DeleteClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{newTenancyFilterClause(f, false)}
}

// tenancyFilterClause implements clause.Interface and gorm.StatementModifier, where gorm.StatementModifier do the real work.
// See gorm.DeletedAt for impl. reference
type tenancyFilterClause struct {
	types.NoopStatementModifier
	Flag  TenancyCheckFlag
	Mode  tcMode
	Field *schema.Field
}

func newTenancyFilterClause(f *schema.Field, isRead bool) *tenancyFilterClause {
	mode := tcMode(TenancyCheckFlagWriteValueCheck)
	tag := extractTenancyFilterTag(f)
	switch tag {
	case "":
		mode = tcModeDefault
	case "-":
	default:
		if strings.ContainsRune(tag, 'r') {
			mode = mode | tcMode(TenancyCheckFlagReadFiltering)
		}
		if strings.ContainsRune(tag, 'w') {
			mode = mode | tcMode(TenancyCheckFlagWriteFiltering)
		}
	}
	flag := TenancyCheckFlagWriteFiltering
	if isRead {
		flag = TenancyCheckFlagReadFiltering
	}
	return &tenancyFilterClause{
		Flag:  flag,
		Mode:  mode,
		Field: f,
	}
}

func extractTenancyFilterTag(f *schema.Field) string {
	if tag, ok := f.Tag.Lookup(types.TagFilter); ok {
		return strings.ToLower(strings.TrimSpace(tag))
	}
	// check if tag is available on embedded Tenancy
	sf, ok := reflectutils.FindStructField(f.Schema.ModelType, func(t reflect.StructField) bool {
		return t.Anonymous && (t.Type.AssignableTo(typeTenancy) || t.Type.AssignableTo(typeTenancyPtr))
	})
	if ok {
		return sf.Tag.Get(types.TagFilter)
	}
	return ""
}

func (c tenancyFilterClause) ModifyStatement(stmt *gorm.Statement) {
	if shouldSkip(stmt.Context, c.Flag, c.Mode) {
		return
	}

	tenantIDs := requiredTenancyFiltering(stmt)
	if len(tenantIDs) == 0 {
		return
	}

	// special fix for db.Model(&model{}).Where(&model{f1:v1}).Or(&model{f2:v2})...
	// Ref:	https://github.com/go-gorm/gorm/issues/3627
	//		https://github.com/go-gorm/gorm/commit/9b2181199d88ed6f74650d73fa9d20264dd134c0#diff-e3e9193af67f3a706b3fe042a9f121d3609721da110f6a585cdb1d1660fd5a3c
	types.FixWhereClausesForStatementModifier(stmt)

	// add tenancy filter condition
	colExpr := stmt.Quote(clause.Column{Table: clause.CurrentTable, Name: c.Field.DBName})
	sql := fmt.Sprintf("%s @> ?", colExpr)
	var conditions []clause.Expression
	for _, id := range tenantIDs {
		conditions = append(conditions, clause.Expr{
			SQL:  sql,
			Vars: []interface{}{UUIDArray{id}},
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
	ta, _ := auth.Details().(securityinternal.TenantAccessDetails)
	if ta != nil {
		idsStr := ta.EffectiveAssignedTenantIds()
		if idsStr.Has(security.SpecialTenantIdWildcard) {
			return nil
		}
		tenantIDs = make([]uuid.UUID, 0, len(idsStr))
		for tenant := range idsStr {
			if tenantId, e := uuid.Parse(tenant); e == nil {
				tenantIDs = append(tenantIDs, tenantId)
			}
		}
	}
	return
}
