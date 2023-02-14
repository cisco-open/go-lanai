package tx

import (
	"context"
	"database/sql"
	"errors"
	"gorm.io/gorm"
)

const (
	ErrTmplSPFailure = `SavePoint failed. did you pass along the context provided by Begin(...)?`
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

// NewGormTxContext will wrap the given Context and *gorm.DB in a gormTxContext
func NewGormTxContext(ctx context.Context, db *gorm.DB) context.Context {
	return gormTxContext{
		txContext: newGormTxContext(ctx),
		db:        db,
	}
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

type TransactionExecuterOptions struct {
	MaxRetries int
}

type TransactionExecuterOption func(options *TransactionExecuterOptions)

// MaxRetries will return a TransactionExecuterOption of type OrderedTransactionExecuterOption
func MaxRetries(maxRetries int, order int) TransactionExecuterOption {
	return func(options *TransactionExecuterOptions) {
		options.MaxRetries = maxRetries
	}
}

type TransactionExecuter interface {
	ExecuteTx(context.Context, *gorm.DB, *sql.TxOptions, TxFunc) error
	Begin(ctx context.Context, db *gorm.DB, opts ...*sql.TxOptions) (context.Context, error)
	Rollback(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) (context.Context, error)
	SavePoint(ctx context.Context, name string) (context.Context, error)
	RollbackTo(ctx context.Context, name string) (context.Context, error)
}

// gormTxManager implements TxManager, ManualTxManager and GormTxManager
type gormTxManager struct {
	db         *gorm.DB
	txExecuter TransactionExecuter
}

func newGormTxManager(db *gorm.DB, executer TransactionExecuter) *gormTxManager {
	return &gormTxManager{
		db:         db,
		txExecuter: executer,
	}
}

func (m gormTxManager) WithDB(db *gorm.DB) GormTxManager {
	return &gormTxManager{
		db:         db,
		txExecuter: m.txExecuter,
	}
}

func (m gormTxManager) Transaction(ctx context.Context, tx TxFunc, opts ...*sql.TxOptions) error {
	// According to finisher_api.go, in the Begin() function, if len(opts) > 0, then it only
	// uses the opts[0] as the option
	var opt *sql.TxOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	return m.txExecuter.ExecuteTx(ctx, m.db, opt, tx)
}

func (m gormTxManager) Begin(ctx context.Context, opts ...*sql.TxOptions) (context.Context, error) {
	// ctx, and get DB out of ctx, and using it here
	return m.txExecuter.Begin(ctx, m.db, opts...)
}

func (m gormTxManager) Rollback(ctx context.Context) (context.Context, error) {
	return m.txExecuter.Rollback(ctx)
}

func (m gormTxManager) Commit(ctx context.Context) (context.Context, error) {
	return m.txExecuter.Commit(ctx)
}

func (m gormTxManager) SavePoint(ctx context.Context, name string) (context.Context, error) {
	return m.txExecuter.SavePoint(ctx, name)
}

func (m gormTxManager) RollbackTo(ctx context.Context, name string) (context.Context, error) {
	return m.txExecuter.RollbackTo(ctx, name)
}

// gormTxManagerAdapter bridge a TxManager to GormTxManager with noop operation. Useful for testing
type gormTxManagerAdapter struct {
	TxManager
}

func (a gormTxManagerAdapter) WithDB(_ *gorm.DB) GormTxManager {
	return a
}

func DoWithDB(ctx context.Context, fn func(*gorm.DB) *gorm.DB) error {
	if gc, ok := ctx.(GormContext); ok {
		if t := gc.DB(); t != nil {
			r := fn(t)
			return r.Error
		}
	}
	return nil
}

// The below code is taken from crdb/tx.go in the crdb package
func ErrIsRetryable(err error) bool {
	// We look for the standard PG errcode SerializationFailureError:40001
	code := errCode(err)
	return code == "40001"
}

func errCode(err error) string {
	var sqlErr errWithSQLState
	if errors.As(err, &sqlErr) {
		return sqlErr.SQLState()
	}

	return ""
}

// errWithSQLState is implemented by pgx (pgconn.PgError) and lib/pq
type errWithSQLState interface {
	SQLState() string
}
