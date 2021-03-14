package tx

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"database/sql"
	"gorm.io/gorm"
)

var (
	ctxKeyGorm = gormCtxKey{}
)

type gormCtxKey struct{}

type gormTxContext struct {
	txContext
	db *gorm.DB
}

func (c gormTxContext) Value(key interface{}) interface{} {
	if k, ok := key.(gormCtxKey); ok && k == ctxKeyGorm {
		return c.db
	}
	return c.Context.Value(key)
}

func GormTxFromContext(ctx context.Context) *gorm.DB {
	if db, ok := ctx.Value(ctxKeyGorm).(*gorm.DB); ok {
		return db.WithContext(ctx)
	}
	return nil
}

// gormTxManager implements TxManager & ManualTxManager
type gormTxManager struct {
	db *gorm.DB
}

func NewGormTxManager(db *gorm.DB) TxManager {
	return &gormTxManager{
		db: db,
	}
}

func (m gormTxManager) Transaction(ctx context.Context, tx TxFunc, opts ...*sql.TxOptions) error {
	return m.db.Transaction(func(txDb *gorm.DB) error {
		c := gormTxContext{
			txContext: txContext{Context: ctx},
			db: txDb,
		}
		return tx(c)
	}, opts...)
}

func (m gormTxManager) Begin(ctx context.Context, opts ...*sql.TxOptions) context.Context {
	tx := m.db.Begin(opts...)
	return gormTxContext{
		txContext: txContext{Context: ctx},
		db: tx,
	}
}

func (m gormTxManager) Rollback(ctx context.Context) context.Context {
	if db, ok := ctx.Value(ctxKeyGorm).(*gorm.DB); ok {
		db.Rollback()
	}

	if orig, ok := ctx.Value(ctxKeyBeginCtx).(context.Context); ok {
		return orig
	}
	panic(data.NewDataError(data.ErrorCodeInvalidTransaction, "rollback failed. did you pass along the context provided by Begin(...)?"))
}

func (m gormTxManager) Commit(ctx context.Context) context.Context {
	if db, ok := ctx.Value(ctxKeyGorm).(*gorm.DB); ok {
		db.Commit()
	}

	if orig, ok := ctx.Value(ctxKeyBeginCtx).(context.Context); ok {
		return orig
	}
	panic(data.NewDataError(data.ErrorCodeInvalidTransaction, "commit failed. did you pass along the context provided by Begin(...)?"))
}

func (m gormTxManager) SavePoint(ctx context.Context, name string) context.Context {
	if db, ok := ctx.Value(ctxKeyGorm).(*gorm.DB); ok {
		db.SavePoint(name)
	}

	if orig, ok := ctx.Value(ctxKeyBeginCtx).(context.Context); ok {
		return orig
	}
	panic(data.NewDataError(data.ErrorCodeInvalidTransaction, "SavePoint failed. did you pass along the context provided by Begin(...)?"))
}

func (m gormTxManager) RollbackTo(ctx context.Context, name string) context.Context {
	if db, ok := ctx.Value(ctxKeyGorm).(*gorm.DB); ok {
		db.RollbackTo(name)
	}

	if orig, ok := ctx.Value(ctxKeyBeginCtx).(context.Context); ok {
		return orig
	}
	panic(data.NewDataError(data.ErrorCodeInvalidTransaction, "SavePoint failed. did you pass along the context provided by Begin(...)?"))
}

