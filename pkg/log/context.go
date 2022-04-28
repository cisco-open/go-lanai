package log

import (
	"context"
	"github.com/go-kit/kit/log"
)

//common fields added by us
const (
	LogKeyMessage   = "msg"
	LogKeyName      = "logger"
	LogKeyTimestamp = "time"
	LogKeyCaller    = "caller"
	LogKeyLevel     = "level"
	LogKeyContext   = "ctx"
)

type ContextValuers map[string]ContextValuer
type ContextValuer func(ctx context.Context) interface{}

type Logger interface {
	log.Logger
	FmtLogger
	KVLogger
	KeyValuer
	Leveler
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
	WithKV(keyvals...interface{}) Logger
}

type KVLogger interface {
	Debug(msg string, keyvals... interface{})
	Info(msg string, keyvals... interface{})
	Warn(msg string, keyvals... interface{})
	Error(msg string, keyvals... interface{})
}

type FmtLogger interface {
	Debugf(msg string, args... interface{})
	Infof(msg string, args... interface{})
	Warnf(msg string, args... interface{})
	Errorf(msg string, args... interface{})
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
	createLogger (name string) ContextualLogger
	addContextValuers(extractors...ContextValuers)
	setLevel (name string, logLevel LoggingLevel)
	refresh (properties *Properties)
}
