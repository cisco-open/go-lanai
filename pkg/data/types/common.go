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

package types

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	TagFilter = "filter"
)

// FixWhereClausesForStatementModifier applies special fix for
// db.Model(&model{}).Where(&model{f1:v1}).Or(&model{f2:v2})...
// Ref:	https://github.com/go-gorm/gorm/issues/3627
//		https://github.com/go-gorm/gorm/commit/9b2181199d88ed6f74650d73fa9d20264dd134c0#diff-e3e9193af67f3a706b3fe042a9f121d3609721da110f6a585cdb1d1660fd5a3c
// Important: utility function for go-lanai internal use
func FixWhereClausesForStatementModifier(stmt *gorm.Statement) {
	cl, _ := stmt.Clauses["WHERE"]
	if where, ok := cl.Expression.(clause.Where); ok && len(where.Exprs) > 1 {
		for _, expr := range where.Exprs {
			if orCond, ok := expr.(clause.OrConditions); ok && len(orCond.Exprs) == 1 {
				where.Exprs = []clause.Expression{clause.And(where.Exprs...)}
				cl.Expression = where
				stmt.Clauses["WHERE"] = cl
				break
			}
		}
	}
}

// NoopStatementModifier used to be embedded of any StatementModifier implementation.
// This type implement dummy clause.Interface methods
type NoopStatementModifier struct {}

func (sm NoopStatementModifier) Name() string {
	// noop
	return ""
}

func (sm NoopStatementModifier) Build(clause.Builder) {
	// noop
}

func (sm NoopStatementModifier) MergeClause(*clause.Clause) {
	// noop
}
