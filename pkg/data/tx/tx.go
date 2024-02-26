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
    "github.com/cisco-open/go-lanai/pkg/data"
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
