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
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gorm.io/gorm"
)

const (
	gormPluginTracing = gormCallbackPrefix + "tracing"
	tracingOpName = "db"
)

type gormConfigurer struct {
	tracer opentracing.Tracer
}

func NewGormTracingConfigurer(tracer opentracing.Tracer) GormConfigurer {
	return &gormConfigurer{
		tracer: tracer,
	}
}

func (c gormConfigurer) Order() int {
	return order.Highest + 1
}

func (c gormConfigurer) Configure(config *gorm.Config) {
	if config.Plugins == nil {
		config.Plugins = map[string]gorm.Plugin{}
	}
	config.Plugins[gormPluginTracing] = &gormPlugin{
		tracer: c.tracer,
	}
}

type gormCallbackFunc func(*gorm.DB)

type gormPlugin struct {
	tracer opentracing.Tracer
}

// Name implements gorm.Plugin
func (p gormPlugin) Name() string {
	return "tracing"
}

// Initialize implements gorm.Plugin. This function register tracing related callbacks
// Default callbacks can be found at github.com/go-gorm/gorm/callbacks/callbacks.go
func (p gormPlugin) Initialize(db *gorm.DB) error {
	_ = db.Callback().Create().Before(GormCallbackBeforeCreate).
		Register(p.cbBeforeName("create"), p.makeBeforeCallback("create"))
	_ = db.Callback().Create().After(GormCallbackAfterCreate).
		Register(p.cbAfterName("create"), p.makeAfterCallback("create"))

	_ = db.Callback().Query().Before(GormCallbackBeforeQuery).
		Register(p.cbBeforeName("query"), p.makeBeforeCallback("select"))
	_ = db.Callback().Query().After(GormCallbackAfterQuery).
		Register(p.cbAfterName("query"), p.makeAfterCallback("select"))

	_ = db.Callback().Update().Before(GormCallbackBeforeUpdate).
		Register(p.cbBeforeName("update"), p.makeBeforeCallback("update"))
	_ = db.Callback().Update().After(GormCallbackAfterUpdate).
		Register(p.cbAfterName("update"), p.makeAfterCallback("update"))

	_ = db.Callback().Delete().Before(GormCallbackBeforeDelete).
		Register(p.cbBeforeName("delete"), p.makeBeforeCallback("delete"))
	_ = db.Callback().Delete().After(GormCallbackAfterDelete).
		Register(p.cbAfterName("delete"), p.makeAfterCallback("delete"))

	_ = db.Callback().Row().Before(GormCallbackBeforeRow).
		Register(p.cbBeforeName("row"), p.makeBeforeCallback("row"))
	_ = db.Callback().Row().After(GormCallbackAfterRow).
		Register(p.cbAfterName("row"), p.makeAfterCallback("row"))

	_ = db.Callback().Raw().Before(GormCallbackBeforeRaw).
		Register(p.cbBeforeName("raw"), p.makeBeforeCallback("sql"))
	_ = db.Callback().Raw().After(GormCallbackAfterRaw).
		Register(p.cbAfterName("raw"), p.makeAfterCallback("sql"))

	return nil
}

func (p gormPlugin) makeBeforeCallback(opName string) gormCallbackFunc {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		name := tracingOpName + " " + opName
		table := db.Statement.Table
		if db.Statement.TableExpr != nil {
			table = db.Statement.TableExpr.SQL
		}

		opts := []tracing.SpanOption{
			tracing.SpanKind(ext.SpanKindRPCClientEnum),
			tracing.SpanTag("table", table),
		}

		db.Statement.Context = tracing.WithTracer(p.tracer).
			WithOpName(name).
			WithOptions(opts...).
			DescendantOrNoSpan(ctx)
	}
}

func (p gormPlugin) makeAfterCallback(_ string) gormCallbackFunc {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		op := tracing.WithTracer(p.tracer)
		if db.Error != nil {
			op = op.WithOptions(tracing.SpanTag("err", db.Error))
		} else {
			op = op.WithOptions(tracing.SpanTag("rows", db.RowsAffected))
		}
		db.Statement.Context = op.FinishAndRewind(ctx)
	}
}

func (p gormPlugin) cbBeforeName(name string) string {
	return gormCallbackPrefix + "before_" + name
}

func (p gormPlugin) cbAfterName(name string) string {
	return gormCallbackPrefix + "after_" + name
}
