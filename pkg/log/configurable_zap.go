package log

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
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

