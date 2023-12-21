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

package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	opalogging "github.com/open-policy-agent/opa/logging"
)

var logger = log.New("OPA")

var (
	logLevelMapper = map[opalogging.Level]log.LoggingLevel{
		opalogging.Debug: log.LevelDebug,
		opalogging.Info:  log.LevelInfo,
		opalogging.Warn:  log.LevelWarn,
		opalogging.Error: log.LevelError,
	}
)

/*******************
	OPA logger
 *******************/

// opaLogger implement logging.Logger
type opaLogger struct {
	logger log.Logger
	level  opalogging.Level
}

func NewOPALogger(logger log.Logger, lvl log.LoggingLevel) opalogging.Logger {
	var level opalogging.Level
	switch lvl {
	case log.LevelDebug:
		level = opalogging.Debug
	case log.LevelWarn:
		level = opalogging.Warn
	case log.LevelError:
		level = opalogging.Error
	default:
		level = opalogging.Info
	}
	return &opaLogger{
		logger: logger.WithLevel(lvl),
		level:  level,
	}
}

func (l *opaLogger) WithContext(ctx context.Context) *opaLogger {
	return &opaLogger{
		logger: logger.WithContext(ctx),
		level:  l.level,
	}
}

func (l *opaLogger) Debug(fmt string, args ...interface{}) {
	l.logger.Debugf(fmt, args...)
}

func (l *opaLogger) Info(fmt string, args ...interface{}) {
	l.logger.Infof(fmt, args...)
}

func (l *opaLogger) Warn(fmt string, args ...interface{}) {
	l.logger.Warnf(fmt, args...)
}

func (l *opaLogger) Error(fmt string, args ...interface{}) {
	l.logger.Errorf(fmt, args...)
}

func (l *opaLogger) WithFields(fields map[string]interface{}) opalogging.Logger {
	kvs := make([]interface{}, 0, 10)
	for k, v := range fields {
		kvs = append(kvs, k, v)
	}
	return &opaLogger{
		logger: l.logger.WithKV(kvs...),
		level:  l.level,
	}
}

func (l *opaLogger) GetLevel() opalogging.Level {
	return l.level
}

func (l *opaLogger) SetLevel(lvl opalogging.Level) {
	newLvl, ok := logLevelMapper[lvl]
	if !ok {
		return
	}
	l.logger = l.logger.WithLevel(newLvl)
	l.level = lvl
}
