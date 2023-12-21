// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package tx

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"database/sql"
	"errors"
	"gorm.io/gorm"
)

var (
	ErrExceededMaxRetries = errors.New("exceeded maximum number of retries")
)

// DefaultExecuter executes the default behaviour for the TxManager and ManualTxManager
// It is possible to define a custom Executer and swap it with this DefaultExecuter.
// To do so, see tx/package.go and the TransactionExecuter in the fx.In of the provideGormTxManager
type DefaultExecuter struct {
	// maxRetries can be defined by using the FxTransactionExecuterOption and MaxRetries option
	maxRetries int
}

func NewDefaultExecuter(options ...TransactionExecuterOption) TransactionExecuter {
	var opts TransactionExecuterOptions
	for _, o := range options {
		o(&opts)
	}
	return &DefaultExecuter{
		maxRetries: opts.MaxRetries,
	}
}

func (r *DefaultExecuter) ExecuteTx(ctx context.Context, db *gorm.DB, opt *sql.TxOptions, txFunc TxFunc) error {
	retryCount := 0

	// if we're in a transaction, make sure to use that db instead
	if gormContext, ok := ctx.(GormContext); ok {
		db = gormContext.DB()
	}
	for {
		err := db.Transaction(func(txDb *gorm.DB) error {
			txErr := txFunc(NewGormTxContext(ctx, txDb)) //nolint:contextcheck // this is equivalent to context.WithXXX
			return txErr
		}, opt)
		if err == nil {
			return nil
		}
		if !ErrIsRetryable(err) {
			return err
		}
		retryCount++
		if r.maxRetries > 0 && retryCount > r.maxRetries {
			return ErrExceededMaxRetries
		}
	}
}

func (r *DefaultExecuter) Begin(ctx context.Context, db *gorm.DB, opts ...*sql.TxOptions) (context.Context, error) {
	//if we're in a transaction, make sure to use that db instead
	if gormContext, ok := ctx.(GormContext); ok {
		db = gormContext.DB()
	}
	tx := db.Begin(opts...)
	if tx.Error != nil {
		return ctx, tx.Error
	}
	return NewGormTxContext(ctx, tx), nil
}

func (r *DefaultExecuter) Rollback(ctx context.Context) (context.Context, error) {
	e := DoWithDB(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Rollback()
	})
	if e != nil {
		return ctx, e
	}

	if tc, ok := ctx.(TxContext); ok && tc.Parent() != nil {
		return tc.Parent(), nil
	}
	return ctx, data.NewDataError(data.ErrorCodeInvalidTransaction, ErrTmplSPFailure)
}

func (r *DefaultExecuter) Commit(ctx context.Context) (context.Context, error) {
	e := DoWithDB(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Commit()
	})
	if e != nil {
		return ctx, e
	}

	if tc, ok := ctx.(TxContext); ok && tc.Parent() != nil {
		return tc.Parent(), nil
	}
	return ctx, data.NewDataError(data.ErrorCodeInvalidTransaction, ErrTmplSPFailure)
}

func (r *DefaultExecuter) SavePoint(ctx context.Context, name string) (context.Context, error) {
	e := DoWithDB(ctx, func(db *gorm.DB) *gorm.DB {
		return db.SavePoint(name)
	})
	if e != nil {
		return ctx, e
	}

	if tc, ok := ctx.(TxContext); ok && tc.Parent() != nil {
		return ctx, nil
	}
	return ctx, data.NewDataError(data.ErrorCodeInvalidTransaction, ErrTmplSPFailure)
}

func (r *DefaultExecuter) RollbackTo(ctx context.Context, name string) (context.Context, error) {
	e := DoWithDB(ctx, func(db *gorm.DB) *gorm.DB {
		return db.RollbackTo(name)
	})
	if e != nil {
		return ctx, e
	}

	if tc, ok := ctx.(TxContext); ok && tc.Parent() != nil {
		return ctx, nil
	}
	return ctx, data.NewDataError(data.ErrorCodeInvalidTransaction, ErrTmplSPFailure)
}
