package tx

import (
	"context"
	"database/sql"
)

type TxFunc func(ctx context.Context) error

type TxManager interface{
	Transaction(ctx context.Context, tx TxFunc, opts ...*sql.TxOptions) error
}

type ManualTxManager interface{
	Begin(ctx context.Context, opts ...*sql.TxOptions) context.Context
	Rollback(ctx context.Context) context.Context
	Commit(ctx context.Context) context.Context
	SavePoint(ctx context.Context, name string) context.Context
	RollbackTo(ctx context.Context, name string) context.Context
}

type txBacktraceCtxKey struct{}

var ctxKeyBeginCtx = txBacktraceCtxKey{}

// txContext helps ManualTxManager to backtrace context used for ManualTxManager.Begin
type txContext struct {
	context.Context
}

func (c txContext) Value(key interface{}) interface{} {
	if k, ok := key.(txBacktraceCtxKey); ok && k == ctxKeyBeginCtx {
		return c.Context
	}
	return c.Context.Value(key)
}