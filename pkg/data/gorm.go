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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"go.uber.org/fx"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"time"
)

const (
	GormCallbackBeforeCreate = "gorm:before_create"
	GormCallbackAfterCreate  = "gorm:after_create"
	GormCallbackBeforeQuery  = "gorm:query"
	GormCallbackAfterQuery   = "gorm:after_query"
	GormCallbackBeforeUpdate = "gorm:before_update"
	GormCallbackAfterUpdate  = "gorm:after_update"
	GormCallbackBeforeDelete = "gorm:before_delete"
	GormCallbackAfterDelete  = "gorm:after_delete"
	GormCallbackBeforeRow    = "gorm:row"
	GormCallbackAfterRow     = "gorm:row"
	GormCallbackBeforeRaw    = "gorm:raw"
	GormCallbackAfterRaw     = "gorm:raw"
)

type gormInitDI struct {
	fx.In
	Dialector gorm.Dialector
	Properties DataProperties
	Configurers [] GormConfigurer `group:"gorm_config"`
}

type GormErrorTranslator interface {
	TranslateWithDB(db *gorm.DB) error
}

type GormConfigurer interface {
	Configure(config *gorm.Config)
}

func NewGorm(di gormInitDI) *gorm.DB {
	level := gormlogger.Warn
	switch di.Properties.Logging.Level {
	case log.LevelOff:
		level = gormlogger.Silent
	case log.LevelDebug, log.LevelInfo:
		level = gormlogger.Info
	case log.LevelWarn:
		level = gormlogger.Warn
	case log.LevelError:
		level = gormlogger.Error
	}

	slow := time.Duration(di.Properties.Logging.SlowThreshold)
	if slow == 0 {
		slow = 15 * time.Second
	}

	config := gorm.Config{
		Logger: newGormLogger(level, slow),
	}

	// gave configurer an chance
	order.SortStable(di.Configurers, order.OrderedFirstCompare)
	for _, c := range di.Configurers {
		c.Configure(&config)
	}

	db, e := gorm.Open(di.Dialector, &config)
	if e != nil {
		panic(e)
	}
	return db
}
