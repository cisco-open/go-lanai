package repo

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"gorm.io/gorm"
)

type GormApi struct {
	db *gorm.DB
}

func (g GormApi) DB(ctx context.Context) *gorm.DB {
	// tx support
	if db := tx.GormTxFromContext(ctx); db != nil {
		return db
	}

	return g.db.WithContext(ctx)
}