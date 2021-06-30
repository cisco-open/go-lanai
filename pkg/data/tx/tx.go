package tx

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"database/sql"
	"go.uber.org/fx"
)

var txManager TxManager

type globalDI struct {
	fx.In
	Tx TxManager `name:"tx/TxManager"`
}

func setGlobalTxManager(di globalDI) {
	txManager = di.Tx
}

func NewTxManager(m TxManager){
	setGlobalTxManager(m)
}

// Transaction start a transaction as a block, return error will rollback, otherwise to commit.
func Transaction(ctx context.Context, tx TxFunc, opts ...*sql.TxOptions) error {
	return mustGetTxManager().Transaction(ctx, tx, opts...)
}

// Begin start a transaction. the returned context.Context should be used for any transactioanl operations
// if returns an error, the returned context.Context should be disgarded
func Begin(ctx context.Context, opts ...*sql.TxOptions) (context.Context, error) {
	return mustGetTxManager().(ManualTxManager).Begin(ctx, opts...)
}

// Rollback rollback a transaction. the returned context.Context is the original provided context when Begin is called
// if returns an error, the returned context.Context should be disgarded
func Rollback(ctx context.Context) (context.Context, error) {
	return mustGetTxManager().(ManualTxManager).Rollback(ctx)
}

// Commit commit a transaction. the returned context.Context is the original provided context when Begin is called
// if returns an error, the returned context.Context should be disgarded
func Commit(ctx context.Context) (context.Context, error) {
	return mustGetTxManager().(ManualTxManager).Commit(ctx)
}

// SavePoint works with RollbackTo and have to be within an transaction.
// the returned context.Context should be used for any transactioanl operations between corresponding SavePoint and RollbackTo
// if returns an error, the returned context.Context should be disgarded
func SavePoint(ctx context.Context, name string) (context.Context, error) {
	return mustGetTxManager().(ManualTxManager).SavePoint(ctx, name)
}

// RollbackTo works with SavePoint and have to be within an transaction.
// the returned context.Context should be used for any transactioanl operations between corresponding SavePoint and RollbackTo
// if returns an error, the returned context.Context should be disgarded
func RollbackTo(ctx context.Context, name string) (context.Context, error) {
	return mustGetTxManager().(ManualTxManager).RollbackTo(ctx, name)
}

func mustGetTxManager() TxManager {
	if txManager == nil {
		panic(data.NewDataError(data.ErrorCodeInternal, "TxManager is not initialized yet. Too early to call tx functions"))
	}
	return txManager
}
