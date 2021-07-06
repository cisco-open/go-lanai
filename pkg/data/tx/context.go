package tx

import (
	"context"
	"database/sql"
)

//goland:noinspection GoNameStartsWithPackageName
type TxFunc func(ctx context.Context) error

//goland:noinspection GoNameStartsWithPackageName
type TxManager interface{
	Transaction(ctx context.Context, tx TxFunc, opts ...*sql.TxOptions) error
}

// ManualTxManager defines interfaces for manual transaction management
// if any methods returns an error, the returned context should be disgarded
type ManualTxManager interface{
	Begin(ctx context.Context, opts ...*sql.TxOptions) (context.Context, error)
	Rollback(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) (context.Context, error)
	SavePoint(ctx context.Context, name string) (context.Context, error)
	RollbackTo(ctx context.Context, name string) (context.Context, error)
}

//goland:noinspection GoNameStartsWithPackageName
type TxContext interface {
	Parent() context.Context
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

func (c txContext) Parent() context.Context {
	return c.Context
}