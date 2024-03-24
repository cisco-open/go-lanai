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
)

//goland:noinspection GoNameStartsWithPackageName
const (
	LogKeyMessage    = internal.LogKeyMessage
	LogKeyName       = internal.LogKeyName
	LogKeyTimestamp  = internal.LogKeyTimestamp
	LogKeyCaller     = internal.LogKeyCaller
	LogKeyLevel      = internal.LogKeyLevel
	LogKeyContext    = internal.LogKeyContext
	LogKeyStacktrace = internal.LogKeyStacktrace
)

type ContextValuers map[string]ContextValuer
type ContextValuer func(ctx context.Context) interface{}

type Logger interface {
	FmtLogger
	KVLogger
	KeyValuer
	Leveler
	CallerValuer
	StdLogger
}

type ContextualLogger interface {
	Logger
	WithContext(ctx context.Context) Logger
}

type Contextual interface {
	WithContext(ctx context.Context) Logger
}

type KeyValuer interface {
	WithKV(keyvals ...interface{}) Logger
}

type CallerValuer interface {
	WithCaller(caller interface{}) Logger
}

type KVLogger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
}

type FmtLogger interface {
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
}

type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type Leveler interface {
	WithLevel(lvl LoggingLevel) Logger
}

type loggerFactory interface {
	createLogger(name string) ContextualLogger
	addContextValuers(valuers ...ContextValuers)
	setRootLevel(logLevel LoggingLevel) (affected int)
	setLevel(prefix string, logLevel *LoggingLevel) (affected int)
	refresh(properties *Properties) error
}
