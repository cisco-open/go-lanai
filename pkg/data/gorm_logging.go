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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	gormlogger "gorm.io/gorm/logger"
	"time"
)

const (
	logKeyDb = "db"
)

type dbLogEntry struct {
	Type        string        `json:"type"`
	TimeElapsed time.Duration `json:"duration"`
	Error       string        `json:"error"`
	Rows        int           `json:"rows"`
	Query       string        `json:"query"`
}

type GormLogger struct {
	level         gormlogger.LogLevel
	slowThreshold time.Duration
	colored       bool
}

func newGormLogger(level gormlogger.LogLevel, slowThreshold time.Duration) *GormLogger {
	return &GormLogger{
		level:         level,
		slowThreshold: slowThreshold,
		colored:       log.IsTerminal(logger),
	}
}

func (l GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &GormLogger{
		level:         level,
		slowThreshold: l.slowThreshold,
		colored:       l.colored,
	}
}

func (l GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	if l.level >= gormlogger.Info {
		logger.WithContext(ctx).Infof(s, i...)
	}
}

func (l GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	if l.level >= gormlogger.Warn {
		logger.WithContext(ctx).Warnf(s, i...)
	}
}

func (l GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	if l.level >= gormlogger.Error {
		logger.WithContext(ctx).Errorf(s, i...)
	}
}

func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	var kvs *dbLogEntry
	var title string
	switch {
	case err != nil && l.level >= gormlogger.Error:
		sql, rows := fc()
		kvs = &dbLogEntry{
			Type:        "error",
			TimeElapsed: elapsed.Truncate(time.Microsecond),
			Error:       err.Error(),
			Rows:        int(rows),
			Query:       sql,
		}
		title = "Error"
	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.level >= gormlogger.Warn:
		sql, rows := fc()
		kvs = &dbLogEntry{
			Type:        "slow",
			TimeElapsed: elapsed.Truncate(time.Microsecond),
			Rows:        int(rows),
			Query:       sql,
		}
		title = "Slow"
	case l.level == gormlogger.Info:
		sql, rows := fc()
		kvs = &dbLogEntry{
			Type:        "sql",
			TimeElapsed: elapsed.Truncate(time.Microsecond),
			Rows:        int(rows),
			Query:       sql,
		}
		title = "SQL"
	default:
		return
	}

	title = "DB " + title
	if l.colored {
		title = gormlogger.Cyan + title + gormlogger.Reset
	}
	logger.WithContext(ctx).WithKV(logKeyDb, kvs).
		Debugf("[%s] %10v | %d Rows | %s | %s", title, kvs.TimeElapsed, kvs.Rows, kvs.Error, kvs.Query)
}
