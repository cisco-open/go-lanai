package repo

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"database/sql"
	"gorm.io/gorm"
)

type TxWithGormFunc func(ctx context.Context, tx *gorm.DB) error

type GormApi interface{
	DB(ctx context.Context) *gorm.DB
	Transaction(ctx context.Context, txFunc TxWithGormFunc, opts ...*sql.TxOptions) error
	WithSession(config *gorm.Session) GormApi
}

type gormApi struct {
	db        *gorm.DB
	txManager tx.GormTxManager
}

func newGormApi(db *gorm.DB, txManager tx.GormTxManager) GormApi {
	return gormApi{
		db: db,
		txManager: txManager.WithDB(db),
	}
}

func (g gormApi) WithSession(config *gorm.Session) GormApi {
	db := g.db.Session(config)
	return gormApi{
		db: db,
		txManager: g.txManager.WithDB(db),
	}
}

func (g gormApi) DB(ctx context.Context) *gorm.DB {
	// tx support
	if t := tx.GormTxWithContext(ctx); t != nil {
		return t
	}

	return g.db.WithContext(ctx)
}

func (g gormApi) Transaction(ctx context.Context, txFunc TxWithGormFunc, opts ...*sql.TxOptions) error {
	return g.txManager.Transaction(ctx, func(c context.Context) error {
		t := tx.GormTxWithContext(c)
		if t == nil {
			return data.NewDataError(data.ErrorCodeInvalidTransaction, "gorm Tx is not found in context")
		}
		return txFunc(c, t)
	}, opts...)
}