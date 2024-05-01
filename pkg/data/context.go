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

package data

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"gorm.io/gorm"
)

const (
	_                        = iota
	ErrorTranslatorOrderGorm // gorm error -> data error
	ErrorTranslatorOrderData // data error -> data error with status code
)

const (
	GormConfigurerGroup  = "gorm_config"
	DatabaseCreatorGroup = "db_creator"
)

// ErrorTranslator redefines web.ErrorTranslator and order.Ordered
// having this redefinition is to break dependency between data and web package
type ErrorTranslator interface {
	order.Ordered
	Translate(ctx context.Context, err error) error
}

type DbCreator interface {
	CreateDatabaseIfNotExist(ctx context.Context, db *gorm.DB) error
}
