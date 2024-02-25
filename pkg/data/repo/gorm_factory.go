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
	"github.com/cisco-open/go-lanai/pkg/data/tx"
	"gorm.io/gorm"
)

type GormFactory struct {
	db *gorm.DB
	txManager tx.GormTxManager
	api GormApi
}

func newGormFactory(db *gorm.DB, txManager tx.GormTxManager) Factory {
	return &GormFactory{
		db: db,
		txManager: txManager,
		api: newGormApi(db, txManager),
	}
}

func (f GormFactory) NewCRUD(model interface{}, options...interface{}) CrudRepository {
	api := f.NewGormApi(options...)
	crud, e := newGormCrud(api, model)
	if e != nil {
		panic(e)
	}

	return crud
}

func (f GormFactory) NewGormApi(options...interface{}) GormApi {
	api := f.api
	for _, v := range options {
		switch opt := v.(type) {
		case gorm.Session:
			api = api.WithSession(&opt)
		case *gorm.Session:
			api = api.WithSession(opt)
		}
	}
	return api
}