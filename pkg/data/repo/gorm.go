package repo

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"database/sql"
	"gorm.io/gorm"
)

type TxWithGormFunc func(ctx context.Context, tx *gorm.DB) error

type GormApi struct {
	db *gorm.DB
	txManager tx.GormTxManager
}

func (g GormApi) DB(ctx context.Context) *gorm.DB {
	// tx support
	if tx := tx.GormTxWithContext(ctx); tx != nil {
		return tx
	}

	return g.db.WithContext(ctx)
}

func (g GormApi) Transaction(ctx context.Context, txFunc TxWithGormFunc, opts ...*sql.TxOptions) error {
	return g.txManager.Transaction(ctx, func(c context.Context) error {
		tx := tx.GormTxWithContext(c)
		if tx != nil {
			return data.NewDataError(data.ErrorCodeInvalidTransaction, "gorm Tx is not found in context")
		}
		return txFunc(c, tx)
	}, opts...)
}