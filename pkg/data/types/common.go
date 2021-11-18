package types

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// fixWhereClausesForStatementModifier applies special fix for
// db.Model(&model{}).Where(&model{f1:v1}).Or(&model{f2:v2})...
// Ref:	https://github.com/go-gorm/gorm/issues/3627
//		https://github.com/go-gorm/gorm/commit/9b2181199d88ed6f74650d73fa9d20264dd134c0#diff-e3e9193af67f3a706b3fe042a9f121d3609721da110f6a585cdb1d1660fd5a3c
func fixWhereClausesForStatementModifier(stmt *gorm.Statement) {
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
