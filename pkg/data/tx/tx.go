package tx

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"database/sql"
)

var txManager TxManager

func setGlobalTxManager(m TxManager) {
	txManager = m
}

func Transaction(ctx context.Context, tx TxFunc, opts ...*sql.TxOptions) error {
	return mustGetTxManager().Transaction(ctx, tx, opts...)
}

func Begin(ctx context.Context, opts ...*sql.TxOptions) context.Context {
	return mustGetTxManager().(ManualTxManager).Begin(ctx, opts...)
}

func Rollback(ctx context.Context) context.Context {
	return mustGetTxManager().(ManualTxManager).Rollback(ctx)
}

func Commit(ctx context.Context) context.Context {
	return mustGetTxManager().(ManualTxManager).Commit(ctx)
}

func SavePoint(ctx context.Context, name string) context.Context {
	return mustGetTxManager().(ManualTxManager).SavePoint(ctx, name)
}

func RollbackTo(ctx context.Context, name string) context.Context {
	return mustGetTxManager().(ManualTxManager).RollbackTo(ctx, name)
}

func mustGetTxManager() TxManager {
	if txManager == nil {
		panic(data.NewDataError(data.ErrorCodeInternal, "TxManager is not initialized yet. Too early to call tx functions"))
	}
	return txManager
}