package log

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
	"github.com/go-kit/kit/log"
)

// common fields added by us
const (
	LogKeyMessage   = internal.LogKeyMessage
	LogKeyName      = internal.LogKeyName
	LogKeyTimestamp = internal.LogKeyTimestamp
	LogKeyCaller    = internal.LogKeyCaller
	LogKeyLevel     = internal.LogKeyLevel
	LogKeyContext   = internal.LogKeyContext
)

type ContextValuers map[string]ContextValuer
type ContextValuer func(ctx context.Context) interface{}

type Logger interface {
	log.Logger
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
	refresh(properties *Properties)
}
