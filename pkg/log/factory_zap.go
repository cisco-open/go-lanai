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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"
)

var zapEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        LogKeyTimestamp,
	LevelKey:       LogKeyLevel,
	NameKey:        LogKeyName,
	CallerKey:      LogKeyCaller,
	FunctionKey:    zapcore.OmitKey,
	MessageKey:     LogKeyMessage,
	StacktraceKey:  LogKeyStacktrace,
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime: func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		// RFC339 with milliseconds
		encoder.AppendString(time.UTC().Format(`2006-01-02T15:04:05.999Z07:00`))
	},
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller: zapcore.ShortCallerEncoder,
}

type zapCoreCreator func(level zap.AtomicLevel) zapcore.Core

type zapLoggerFactory struct {
	rootLogLevel     LoggingLevel
	logLevels        map[string]LoggingLevel
	coreCreator      zapCoreCreator
	properties       *Properties
	effectiveValuers ContextValuers
	extraValuers     ContextValuers
	registry         map[string]*configurableZapLogger
}

func newZapLoggerFactory(properties *Properties) *zapLoggerFactory {
	rootLogLevel, ok := properties.Levels[keyLevelDefault]
	if !ok {
		rootLogLevel = LevelInfo
	}

	var e error
	f := &zapLoggerFactory{
		rootLogLevel:     rootLogLevel,
		logLevels:        convertLevelsNameToKey(properties.Levels),
		properties:       properties,
		registry:         map[string]*configurableZapLogger{},
		extraValuers:     ContextValuers{},
		effectiveValuers: ContextValuers{},
	}
	f.effectiveValuers = f.buildContextValuer(properties)
	if f.coreCreator, e = f.buildZapCoreCreator(properties); e != nil {
		panic(e)
	}
	return f
}

func (f *zapLoggerFactory) createLogger(name string) ContextualLogger {
	key := loggerKey(name)
	if l, ok := f.registry[key]; ok {
		return l
	}

	ll := f.resolveEffectiveLevel(key)
	leveler := zap.NewAtomicLevel()
	l := newConfigurableZapLogger(name, f.coreCreator(leveler), ll, leveler, f.effectiveValuers)
	f.registry[key] = l
	return l
}

func (f *zapLoggerFactory) addContextValuers(valuers ...ContextValuers) {
	for _, item := range valuers {
		for k, v := range item {
			f.effectiveValuers[k] = v
			f.extraValuers[k] = v
		}
	}
}

func (f *zapLoggerFactory) setRootLevel(logLevel LoggingLevel) (affected int) {
	f.rootLogLevel = logLevel
	for k, l := range f.registry {
		effective := f.resolveEffectiveLevel(k)
		l.setMinLevel(effective)
		affected++
	}
	return
}

func (f *zapLoggerFactory) setLevel(prefix string, logLevel *LoggingLevel) (affected int) {
	key := loggerKey(prefix)
	if (key == "" || key == keyLevelDefault || key == loggerKey(nameLevelDefault)) && logLevel != nil {
		return f.setRootLevel(*logLevel)
	}

	if logLevel == nil {
		// unset
		if _, ok := f.logLevels[key]; ok {
			delete(f.logLevels, key)
		}
	} else {
		// set
		f.logLevels[key] = *logLevel
	}

	// set effective level to all affected loggers
	withDot := key + keySeparator
	for k, l := range f.registry {
		if k != key && !strings.HasPrefix(k, withDot) {
			continue
		}
		effective := f.resolveEffectiveLevel(k)
		l.setMinLevel(effective)
		affected++
	}
	return
}

func (f *zapLoggerFactory) refresh(properties *Properties) error {
	rootLogLevel, ok := properties.Levels[keyLevelDefault]
	if !ok {
		rootLogLevel = LevelInfo
	}

	f.rootLogLevel = rootLogLevel
	f.logLevels = convertLevelsNameToKey(properties.Levels)
	f.effectiveValuers = buildContextValuerFromConfig(properties)
	var e error
	if f.coreCreator, e = f.buildZapCoreCreator(properties); e != nil {
		return e
	}

	// merge valuers, note: we don't delete extra valuers during refresh
	for k, v := range f.extraValuers {
		f.effectiveValuers[k] = v
	}

	for key, l := range f.registry {
		ll := f.resolveEffectiveLevel(key)
		l.core = f.coreCreator(l.leveler)
		l.valuers = f.effectiveValuers
		l.setMinLevel(ll)
	}
	return nil
}

func (f *zapLoggerFactory) resolveEffectiveLevel(key string) LoggingLevel {
	prefix := key
	for i := len(key); i > 0; i = strings.LastIndex(prefix, keySeparator) {
		prefix = key[0:i]
		if ll, ok := f.logLevels[prefix]; ok {
			return ll
		}
	}
	return f.rootLogLevel
}

func (f *zapLoggerFactory) buildContextValuer(properties *Properties) ContextValuers {
	valuers := ContextValuers{}
	// k is context-key, v is log-key
	for k, v := range properties.Mappings {
		valuers[v] = func(ctx context.Context) interface{} {
			return ctx.Value(k)
		}
	}

	return valuers
}

func (f *zapLoggerFactory) buildZapCoreCreator(properties *Properties) (zapCoreCreator, error) {
	if len(properties.Loggers) == 0 {
		properties.Loggers = map[string]*LoggerProperties{
			"default": {
				Type:      TypeConsole,
				Format:    FormatText,
				Template:  defaultTemplate,
				FixedKeys: defaultFixedFields.Values(),
			},
		}
	}
	encoders := make([]zapcore.Encoder, len(properties.Loggers))
	syncers := make([]zapcore.WriteSyncer, len(properties.Loggers))
	var i int
	for _, loggerProps := range properties.Loggers {
		var e error
		if syncers[i], e = f.newZapWriteSyncer(loggerProps); e != nil {
			return nil, e
		}
		if encoders[i], e = f.newZapEncoder(loggerProps, syncers[i].(internal.TerminalAware).IsTerminal()); e != nil {
			return nil, e
		}
		i++
	}
	return func(level zap.AtomicLevel) zapcore.Core {
		var core zapcore.Core
		var isTerm bool
		switch len(encoders) {
		case 0:
			// not possible
			return zapcore.NewNopCore()
		case 1:
			core = zapcore.NewCore(encoders[0], syncers[0], level)
			isTerm = syncers[0].(internal.TerminalAware).IsTerminal()
		default:
			cores := make([]zapcore.Core, len(encoders))
			for i := range encoders {
				cores[i] = zapcore.NewCore(encoders[i], syncers[i], level)
				isTerm = isTerm && syncers[i].(internal.TerminalAware).IsTerminal()
			}
			core = zapcore.NewTee(cores...)
		}
		if isTerm {
			return internal.ZapTerminalCore{Core: core}
		}
		return core
	}, nil
}

func (f *zapLoggerFactory) newZapEncoder(props *LoggerProperties, isTerm bool) (zapcore.Encoder, error) {
	switch props.Format {
	case FormatText:
		fixedFields := defaultFixedFields.Copy().Add(props.FixedKeys...)
		formatter := internal.NewTemplatedFormatter(props.Template, fixedFields, isTerm)
		return internal.NewZapFormattedEncoder(zapEncoderConfig, formatter, isTerm), nil
	case FormatJson:
		return zapcore.NewJSONEncoder(zapEncoderConfig), nil
	}
	return nil, fmt.Errorf("unsupported logger format: %v", props.Format)
}

func (f *zapLoggerFactory) newZapWriteSyncer(props *LoggerProperties) (zapcore.WriteSyncer, error) {
	switch props.Type {
	case TypeConsole:
		return internal.NewZapWriterWrapper(os.Stdout), nil
	case TypeFile:
		file, e := openOrCreateFile(props.Location)
		if e != nil {
			return nil, e
		}
		return internal.NewZapWriterWrapper(file), nil
	case TypeHttp:
		fallthrough
	case TypeMQ:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported logger type: %v", props.Type)
	}
}

