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

package repo

import (
    "context"
    "database/sql"
    "github.com/cisco-open/go-lanai/pkg/data"
    "github.com/cisco-open/go-lanai/pkg/data/tx"
    "gorm.io/gorm"
)

type TxWithGormFunc func(ctx context.Context, tx *gorm.DB) error

type GormApi interface {
	DB(ctx context.Context) *gorm.DB
	Transaction(ctx context.Context, txFunc TxWithGormFunc, opts ...*sql.TxOptions) error
	WithSession(config *gorm.Session) GormApi
}

type gormApi struct {
	db        *gorm.DB
	txManager tx.GormTxManager
}

func newGormApi(db *gorm.DB, txManager tx.GormTxManager) GormApi {
	return gormApi{
		db:        db,
		txManager: txManager.WithDB(db),
	}
}

func (g gormApi) WithSession(config *gorm.Session) GormApi {
	db := g.db.Session(config)
	return gormApi{
		db:        db,
		txManager: g.txManager.WithDB(db),
	}
}

func (g gormApi) DB(ctx context.Context) *gorm.DB {
	// tx support
	if t := tx.GormTxWithContext(ctx); t != nil {
		return t
	}

	return g.db.WithContext(ctx)
}

func (g gormApi) Transaction(ctx context.Context, txFunc TxWithGormFunc, opts ...*sql.TxOptions) error {
	return g.txManager.Transaction(ctx, func(c context.Context) error {
		t := tx.GormTxWithContext(c)
		if t == nil {
			return data.NewDataError(data.ErrorCodeInvalidTransaction, "gorm Tx is not found in context")
		}
		return txFunc(c, t)
	}, opts...)
}
