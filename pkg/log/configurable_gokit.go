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
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"time"
)

var (
	timestampUTC = log.TimestampFormat(
		func() time.Time { return time.Now().UTC() },
		"2006-01-02T15:04:05.999Z07:00",
	)
)

// configurableKitLogger implements Logger and Contextual
type configurableKitLogger struct {
	kitLogger
	name      string
	lvl       LoggingLevel
	template  log.Logger
	swappable *log.SwapLogger
	valuers   ContextValuers
	isTerm    bool
}

func newConfigurableLogger(name string, templateLogger log.Logger, logLevel LoggingLevel, valuers ContextValuers) *configurableKitLogger {
	swap := log.SwapLogger{}
	k := &configurableKitLogger{
		kitLogger: kitLogger{
			Logger: log.With(&swap,
				LogKeyTimestamp, timestampUTC,
				LogKeyCaller, log.Caller(4),
				LogKeyName, name,
			),
		},
		name:      name,
		swappable: &swap,
		template:  templateLogger,
		valuers:   valuers,
		isTerm:    isTerminal(templateLogger),
	}
	k.setLevel(logLevel)
	return k
}

func (l *configurableKitLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	fields := make([]interface{}, 0, len(l.valuers)*2)
	for k, ctxValuer := range l.valuers {
		valuer := makeValuer(ctx, ctxValuer)
		fields = append(fields, k, valuer)
	}
	return l.withKV(fields)
}

func (l *configurableKitLogger) setLevel(lv LoggingLevel) {
	var opt level.Option
	switch lv {
	case LevelOff:
		opt = level.AllowNone()
	case LevelDebug:
		opt = level.AllowDebug()
	case LevelError:
		opt = level.AllowError()
	case LevelInfo:
		opt = level.AllowInfo()
	case LevelWarn:
		opt = level.AllowWarn()
	default:
		opt = level.AllowInfo()
	}
	l.lvl = lv
	l.swappable.Swap(level.NewFilter(l.template, opt))
}

// IsTerminal implements internal.TerminalAware
func (l *configurableKitLogger) IsTerminal() bool {
	return l.isTerm
}

func isTerminal(l log.Logger) bool {
	switch l.(type) {
	case *internal.KitTextLoggerAdapter:
		return l.(*internal.KitTextLoggerAdapter).IsTerminal
	case *compositeKitLogger:
		for _, l := range l.(*compositeKitLogger).delegates {
			if !isTerminal(l) {
				return false
			}
		}
		return true
	case *configurableKitLogger:
		return l.(*configurableKitLogger).isTerm
	default:
		return false
	}
}

func makeValuer(ctx context.Context, ctxValuer ContextValuer) log.Valuer {
	return func() interface{} {
		return ctxValuer(ctx)
	}
}
