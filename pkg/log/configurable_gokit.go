package log

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
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
