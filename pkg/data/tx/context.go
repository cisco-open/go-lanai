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
	"database/sql"
)

//goland:noinspection GoNameStartsWithPackageName
type TxFunc func(ctx context.Context) error

//goland:noinspection GoNameStartsWithPackageName
type TxManager interface {
	Transaction(ctx context.Context, tx TxFunc, opts ...*sql.TxOptions) error
}

// ManualTxManager defines interfaces for manual transaction management
// if any methods returns an error, the returned context should be disgarded
type ManualTxManager interface {
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

// newGormTxContext will check if the given context.Context is a TxContext. If so,
// It will increment the nestLevel of the new TxContext.
func newGormTxContext(ctx context.Context) txContext {
	return txContext{
		Context: ctx,
	}
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
