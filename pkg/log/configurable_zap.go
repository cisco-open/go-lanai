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

package log

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/log/internal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// configurableZapLogger implements Logger and Contextual
type configurableZapLogger struct {
	zapLogger
	leveler zap.AtomicLevel
	lvl     LoggingLevel
	valuers ContextValuers
}

func newConfigurableZapLogger(name string, core zapcore.Core, logLevel LoggingLevel, leveler zap.AtomicLevel, valuers ContextValuers) *configurableZapLogger {
	l := &configurableZapLogger{
		zapLogger: zapLogger{
			core:        core,
			name:        name,
			clock:       zapcore.DefaultClock,
			stacktracer: RuntimeCaller(4),
			fields:      nil,
		},
		leveler: leveler,
		valuers: valuers,
	}
	l.setMinLevel(logLevel)
	return l
}

func (l *configurableZapLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	fields := make([]interface{}, 0, len(l.valuers)*2)
	for k, ctxValuer := range l.valuers {
		fields = append(fields, k, ctxValuer(ctx))
	}
	return l.withKV(fields)
}

func (l *configurableZapLogger) setMinLevel(lv LoggingLevel) {
	switch lv {
	case LevelOff:
		l.leveler.SetLevel(zapcore.InvalidLevel)
	case LevelDebug:
		l.leveler.SetLevel(zapcore.DebugLevel)
	case LevelInfo:
		l.leveler.SetLevel(zapcore.InfoLevel)
	case LevelWarn:
		l.leveler.SetLevel(zapcore.WarnLevel)
	case LevelError:
		l.leveler.SetLevel(zapcore.ErrorLevel)
	default:
		l.leveler.SetLevel(zapcore.InfoLevel)
	}
	l.lvl = lv
}

// IsTerminal implements internal.TerminalAware
func (l *configurableZapLogger) IsTerminal() bool {
	termAware, ok := l.core.(internal.TerminalAware)
	return ok && termAware.IsTerminal()
}

