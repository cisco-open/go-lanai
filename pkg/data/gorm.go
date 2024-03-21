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
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
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

const (
	gormCallbackPrefix = "lanai:"
)

type GormErrorTranslator interface {
	TranslateWithDB(db *gorm.DB) error
}

type GormConfigurer interface {
	Configure(config *gorm.Config)
}

type GormOptions func(cfg *GormConfig)
type GormConfig struct {
	Dialector             gorm.Dialector
	LogLevel              log.LoggingLevel
	LogSlowQueryThreshold time.Duration
	Configurers           []GormConfigurer
}

func NewGorm(opts ...GormOptions) *gorm.DB {
	cfg := GormConfig{
		LogSlowQueryThreshold: 15 * time.Second,
	}
	for _, fn := range opts {
		fn(&cfg)
	}
	level := gormlogger.Warn
	switch cfg.LogLevel {
	case log.LevelOff:
		level = gormlogger.Silent
	case log.LevelDebug, log.LevelInfo:
		level = gormlogger.Info
	case log.LevelWarn:
		level = gormlogger.Warn
	case log.LevelError:
		level = gormlogger.Error
	}

	config := gorm.Config{
		Logger: newGormLogger(level, cfg.LogSlowQueryThreshold),
	}

	// gave configurer an chance
	order.SortStable(cfg.Configurers, order.OrderedFirstCompare)
	for _, c := range cfg.Configurers {
		c.Configure(&config)
	}

	db, e := gorm.Open(cfg.Dialector, &config)
	if e != nil {
		panic(e)
	}
	return db
}
