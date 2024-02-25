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

package dbtest

import (
    "context"
    "database/sql"
    "github.com/cisco-open/go-lanai/pkg/data/tx"
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
