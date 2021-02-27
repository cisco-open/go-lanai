package log

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// configurableLogger implements Logger and Contextual
type configurableLogger struct {
	logger
	name      string
	template  log.Logger
	swappable *log.SwapLogger
	valuers   ContextValuers
}

func newConfigurableLogger(name string, templateLogger log.Logger, logLevel LoggingLevel, valuers ContextValuers) *configurableLogger {
	swap := log.SwapLogger{}
	k := &configurableLogger{
		logger: logger{
			Logger: log.WithPrefix(&swap, LogKeyName, name),
		},
		name:      name,
		swappable: &swap,
		template:  templateLogger,
		valuers:   valuers,
	}
	k.setLevel(logLevel)
	return k
}

func (l *configurableLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	fields := []interface{}{LogKeyContext, ctx}
	for k, ctxValuer := range l.valuers {
		var valuer log.Valuer = func() interface{} {
			return ctxValuer(ctx)
		}
		fields = append(fields, k, valuer)
	}
	return l.withKV(fields)
}

func (l *configurableLogger) setLevel(lv LoggingLevel) {
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
	l.swappable.Swap(level.NewFilter(l.template, opt))
}
