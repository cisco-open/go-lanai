package tx

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"database/sql"
	"gorm.io/gorm"
)

const (
	errTmplSPFailure = `SavePoint failed. did you pass along the context provided by Begin(...)?`
)

type GormTxManager interface {
	TxManager
	WithDB(*gorm.DB) GormTxManager
}

type GormContext interface {
	DB() *gorm.DB
}

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
	return c.txContext.Value(key)
}

func (c gormTxContext) DB() *gorm.DB {
	return c.db
}

func GormTxWithContext(ctx context.Context) (tx *gorm.DB) {
	if c, ok := ctx.(GormContext); ok && c.DB() != nil {
		return c.DB().WithContext(ctx)
	}

	if db, ok := ctx.Value(ctxKeyGorm).(*gorm.DB); ok {
		return db.WithContext(ctx)
	}
	return nil
}

// gormTxManager implements TxManager, ManualTxManager and GormTxManager
type gormTxManager struct {
	db *gorm.DB
}

func newGormTxManager(db *gorm.DB) *gormTxManager {
	return &gormTxManager{
		db: db,
	}
}

func (m gormTxManager) WithDB(db *gorm.DB) GormTxManager {
	return &gormTxManager{
		db: db,
	}
}

func (m gormTxManager) Transaction(ctx context.Context, tx TxFunc, opts ...*sql.TxOptions) error {
	return m.db.Transaction(func(txDb *gorm.DB) error {
		c := gormTxContext{
			txContext: txContext{Context: ctx},
			db:        txDb,
		}
		return tx(c)
	}, opts...)
}

func (m gormTxManager) Begin(ctx context.Context, opts ...*sql.TxOptions) (context.Context, error) {
	tx := m.db.Begin(opts...)
	if tx.Error != nil {
		return ctx, tx.Error
	}
	return gormTxContext{
		txContext: txContext{Context: ctx},
		db:        tx,
	}, nil
}

func (m gormTxManager) Rollback(ctx context.Context) (context.Context, error) {
	e := doWithDB(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Rollback()
	})
	if e != nil {
		return ctx, e
	}

	if tc, ok := ctx.(TxContext); ok && tc.Parent() != nil {
		return tc.Parent(), nil
	}
	return ctx, data.NewDataError(data.ErrorCodeInvalidTransaction, errTmplSPFailure)
}

func (m gormTxManager) Commit(ctx context.Context) (context.Context, error) {
	e := doWithDB(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Commit()
	})
	if e != nil {
		return ctx, e
	}

	if tc, ok := ctx.(TxContext); ok && tc.Parent() != nil {
		return tc.Parent(), nil
	}
	return ctx, data.NewDataError(data.ErrorCodeInvalidTransaction, errTmplSPFailure)
}

func (m gormTxManager) SavePoint(ctx context.Context, name string) (context.Context, error) {
	e := doWithDB(ctx, func(db *gorm.DB) *gorm.DB {
		return db.SavePoint(name)
	})
	if e != nil {
		return ctx, e
	}

	if tc, ok := ctx.(TxContext); ok && tc.Parent() != nil {
		return ctx, nil
	}
	return ctx, data.NewDataError(data.ErrorCodeInvalidTransaction, errTmplSPFailure)
}

func (m gormTxManager) RollbackTo(ctx context.Context, name string) (context.Context, error) {
	e := doWithDB(ctx, func(db *gorm.DB) *gorm.DB {
		return db.RollbackTo(name)
	})
	if e != nil {
		return ctx, e
	}

	if tc, ok := ctx.(TxContext); ok && tc.Parent() != nil {
		return ctx, nil
	}
	return ctx, data.NewDataError(data.ErrorCodeInvalidTransaction, errTmplSPFailure)
}

// gormTxManagerAdapter bridge a TxManager to GormTxManager with noop operation. Useful for testing
type gormTxManagerAdapter struct {
	TxManager
}

func (a gormTxManagerAdapter) WithDB(_ *gorm.DB) GormTxManager {
	return a
}

func doWithDB(ctx context.Context, fn func(*gorm.DB) *gorm.DB) error {
	if gc, ok := ctx.(GormContext); ok {
		if t := gc.DB(); t != nil {
			r := fn(t)
			return r.Error
		}
	}
	return nil
}