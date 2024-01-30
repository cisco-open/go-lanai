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
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime"
)

/************************
	level enabler
 ************************/

type zapLevel LoggingLevel

func (lvl zapLevel) Enabled(zapLvl zapcore.Level) bool {
	switch LoggingLevel(lvl) {
	case LevelDebug:
		return zapLvl >= zapcore.DebugLevel
	case LevelInfo:
		return zapLvl >= zapcore.InfoLevel
	case LevelWarn:
		return zapLvl >= zapcore.WarnLevel
	case LevelError:
		return zapLvl >= zapcore.ErrorLevel
	case LevelOff:
		return false
	default:
		return zapLvl >= 0
	}
}

/************************
	logger
 ************************/

// zapLogger implements Logger
type zapLogger struct {
	core        zapcore.Core
	name        string
	clock       zapcore.Clock
	stacktracer Stacktracer
	fields      []zapcore.Field
	//zap.SugaredLogger
}

func (l *zapLogger) WithKV(keyvals ...interface{}) Logger {
	return l.withKV(keyvals)
}

func (l *zapLogger) WithLevel(lvl LoggingLevel) Logger {
	var leveled zapcore.Core
	var e error
	if leveled, e = zapcore.NewIncreaseLevelCore(l.core, zapLevel(lvl)); e != nil {
		// probably trying to decrease level, use noop
		leveled = zapcore.NewNopCore()
	}
	cpy := l.shallowCopy()
	cpy.core = leveled
	return cpy
}

// WithCaller implements CallerValuer
func (l *zapLogger) WithCaller(caller interface{}) Logger {
	cpy := l.shallowCopy()
	switch fn := caller.(type) {
	case nil:
		cpy.stacktracer = nil
	case func() ([]*runtime.Frame, interface{}):
		cpy.stacktracer = fn
	case Stacktracer:
		cpy.stacktracer = fn
	case func() interface{}:
		cpy.stacktracer = func() (frames []*runtime.Frame, fallback interface{}) {
			return nil, fn()
		}
	default:
		cpy.stacktracer = func() (frames []*runtime.Frame, fallback interface{}) {
			return nil, caller
		}
	}
	return cpy
}

func (l *zapLogger) Debugf(msg string, args ...interface{}) {
	l.log(zapcore.DebugLevel, msg, args, nil)
}

func (l *zapLogger) Infof(msg string, args ...interface{}) {
	l.log(zapcore.InfoLevel, msg, args, nil)
}

func (l *zapLogger) Warnf(msg string, args ...interface{}) {
	l.log(zapcore.WarnLevel, msg, args, nil)
}

func (l *zapLogger) Errorf(msg string, args ...interface{}) {
	l.log(zapcore.ErrorLevel, msg, args, nil)
}

func (l *zapLogger) Debug(msg string, keyvals ...interface{}) {
	l.log(zapcore.DebugLevel, msg, nil, keyvals)
}

func (l *zapLogger) Info(msg string, keyvals ...interface{}) {
	l.log(zapcore.InfoLevel, msg, nil, keyvals)
}

func (l *zapLogger) Warn(msg string, keyvals ...interface{}) {
	l.log(zapcore.WarnLevel, msg, nil, keyvals)
}

func (l *zapLogger) Error(msg string, keyvals ...interface{}) {
	l.log(zapcore.ErrorLevel, msg, nil, keyvals)
}

func (l *zapLogger) Print(args ...interface{}) {
	l.log(zapcore.LevelOf(l.core), "", args, nil)
}

func (l *zapLogger) Printf(format string, args ...interface{}) {
	l.log(zapcore.LevelOf(l.core), format, args, nil)
}

func (l *zapLogger) Println(args ...interface{}) {
	l.log(zapcore.LevelOf(l.core), "\n", args, nil)
}

func (l *zapLogger) Log(keyvals ...interface{}) error {
	l.log(zapcore.LevelOf(l.core), "", nil, keyvals)
	return nil
}

func (l *zapLogger) shallowCopy() *zapLogger {
	cpy := *l
	return &cpy
}

func (l *zapLogger) withKV(keyvals []interface{}) Logger {
	cpy := l.shallowCopy()
	cpy.fields = append(l.fields, l.toFields(keyvals)...)
	return cpy
}

func (l *zapLogger) log(lvl zapcore.Level, msgTmpl string, fmtArgs []interface{}, keyvals []interface{}) {
	// If logging at this level is completely disabled, skip the overhead of string formatting.
	if lvl < zapcore.DPanicLevel && !l.core.Enabled(lvl) {
		return
	}
	msg := l.constructMessage(msgTmpl, fmtArgs)
	ce := l.core.Check(zapcore.Entry{
		LoggerName: l.name,
		Time:       l.clock.Now(),
		Level:      lvl,
		Message:    msg,
	}, nil)
	if ce == nil {
		return
	}
	// process fields
	adhocFields := l.toFields(keyvals)

	// caller and stacktrace
	if l.stacktracer != nil {
		switch frames, fallback := l.stacktracer(); {
		case len(frames) != 0:
			ce.Caller = zapcore.EntryCaller{
				Defined:  frames[0].PC != 0,
				PC:       frames[0].PC,
				File:     frames[0].File,
				Line:     frames[0].Line,
				Function: frames[0].Function,
			}
			//Note: we currently don't support stacktrace. Here is the place to add it
		case fallback != nil:
			adhocFields = append(adhocFields, zap.Any(LogKeyCaller, fallback))
		default:
			// no caller info
		}
	}

	// write log
	ce.Write(append(l.fields, adhocFields...)...)
}

// constructMessage similar to SugarLogger's getMessage(...)
func (l *zapLogger) constructMessage(tmpl string, fmtArgs []interface{}) string {
	switch {
	case len(fmtArgs) == 0:
		return tmpl
	case tmpl == "\n":
		return fmt.Sprintln(fmtArgs...)
	case len(tmpl) != 0:
		return fmt.Sprintf(tmpl, fmtArgs...)
	default:
		return fmt.Sprint(fmtArgs...)
	}
}

func (l *zapLogger) toFields(keyvals []interface{}) []zapcore.Field {
	if len(keyvals) == 0 {
		return nil
	}
	// give it enough space to avoid re-allocate space
	fields := make([]zapcore.Field, 0, len(keyvals)/2+2)
	for i := 0; i < len(keyvals); i += 2 {
		var key string
		switch k := keyvals[i].(type) {
		case string:
			key = k
		case fmt.Stringer:
			key = k.String()
		default:
			key = fmt.Sprint(k)
		}
		if i == len(keyvals)-1 {
			fields = append(fields, zap.String(key, "!(MISSING)"))
			break
		}
		switch v := keyvals[i+1].(type) {
		case zapcore.Field:
			fields = append(fields, v)
		case *zapcore.Field:
			fields = append(fields, *v)
		case func() interface{}:
			fields = append(fields, zap.Any(key, v()))
		default:
			fields = append(fields, zap.Any(key, v))
		}
	}
	return fields
}
