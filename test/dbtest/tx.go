package dbtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"database/sql"
	"gorm.io/gorm"
)

type noopTxManager struct {}

func provideNoopTxManager() tx.TxManager {
	return noopTxManager{}
}

func (m noopTxManager) Transaction(ctx context.Context, fn tx.TxFunc, _ ...*sql.TxOptions) error {
	return fn(m.mockTxContext(ctx))
}

func (m noopTxManager) WithDB(_ *gorm.DB) tx.GormTxManager {
	return m
}

func (m noopTxManager) Begin(ctx context.Context, _ ...*sql.TxOptions) (context.Context, error) {
	return m.mockTxContext(ctx), nil
}

func (m noopTxManager) Rollback(ctx context.Context) (context.Context, error) {
	if tc, ok := ctx.(tx.TxContext); ok {
		return tc.Parent(), nil
	}
	return ctx, nil
}

func (m noopTxManager) Commit(ctx context.Context) (context.Context, error) {
	if tc, ok := ctx.(tx.TxContext); ok {
		return tc.Parent(), nil
	}
	return ctx, nil
}

func (m noopTxManager) SavePoint(ctx context.Context, _ string) (context.Context, error) {
	return ctx, nil
}

func (m noopTxManager) RollbackTo(ctx context.Context, _ string) (context.Context, error) {
	return ctx, nil
}

func (m noopTxManager) mockTxContext(ctx context.Context) context.Context {
	return &mockedGormContext{
		mockedTxContext: mockedTxContext{
			Context: ctx,
		},
		db: &gorm.DB{
			Config:       &gorm.Config{},
			Statement: &gorm.Statement{},
		},
	}
}
