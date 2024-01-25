package log

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

/************************
	compositeKitLogger
 ************************/
type compositeKitLogger struct {
	delegates []log.Logger
}

func (c *compositeKitLogger) addLogger(l log.Logger) {
	c.delegates = append(c.delegates, l)
}

func (c *compositeKitLogger) Log(keyvals ...interface{}) error {
	for _, d := range c.delegates {
		_ = d.Log(keyvals...)
	}
	return nil
}

/************************
	logger
 ************************/

// kitLogger implements Logger
type kitLogger struct {
	log.Logger
}

func (l *kitLogger) WithKV(keyvals...interface{}) Logger {
	return l.withKV(keyvals)
}

func (l *kitLogger) WithLevel(lvl LoggingLevel) Logger {
	var leveled log.Logger
	switch lvl {
	case LevelDebug:
		leveled = level.Debug(l.Logger)
	case LevelInfo:
		leveled = level.Info(l.Logger)
	case LevelWarn:
		leveled = level.Warn(l.Logger)
	case LevelError:
		leveled = level.Error(l.Logger)
	case LevelOff:
		leveled = log.NewNopLogger()
	default:
		return l
	}

	return &kitLogger{
		Logger: leveled,
	}
}

func (l *kitLogger) WithCaller(caller interface{}) Logger {
	switch fn := caller.(type) {
	case func() interface{}:
		// untyped function
		caller = log.Valuer(fn)
	}
	return &kitLogger{
		Logger: log.WithSuffix(l.Logger, LogKeyCaller, caller),
	}
}

func (l *kitLogger) Debugf(msg string, args ...interface{}) {
	_ = l.debugLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *kitLogger) Infof(msg string, args ...interface{}) {
	_ = l.infoLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *kitLogger) Warnf(msg string, args ...interface{}) {
	_ = l.warnLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *kitLogger) Errorf(msg string, args ...interface{}) {
	_ = l.errorLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *kitLogger) Debug(msg string, keyvals... interface{}) {
	_ = l.debugLogger(keyvals).Log(LogKeyMessage, msg)
}

func (l *kitLogger) Info(msg string, keyvals... interface{}) {
	_ = l.infoLogger(keyvals).Log(LogKeyMessage, msg)
}

func (l *kitLogger) Warn(msg string, keyvals... interface{}) {
	_ = l.warnLogger(keyvals).Log(LogKeyMessage, msg)
}

func (l *kitLogger) Error(msg string, keyvals... interface{}) {
	_ = l.errorLogger(keyvals).Log(LogKeyMessage, msg)
}


func (l *kitLogger) Print(args ...interface{})  {
	_ = l.Log(LogKeyMessage, fmt.Sprint(args...))
}

func (l *kitLogger) Printf(format string, args ...interface{}) {
	_ = l.Log(LogKeyMessage, fmt.Sprintf(format, args...))
}

func (l *kitLogger) Println(args ...interface{}) {
	_ = l.Log(LogKeyMessage, fmt.Sprintln(args...))
}

func (l *kitLogger) withKV(keyvals []interface{}) Logger {
	return &kitLogger{
		Logger: log.WithPrefix(l.Logger, keyvals...),
	}
}

func (l *kitLogger) debugLogger(keyvals []interface{}) log.Logger {
	return log.WithPrefix(level.Debug(l.Logger), keyvals...)
}

func (l *kitLogger) infoLogger(keyvals []interface{}) log.Logger {
	return log.WithPrefix(level.Info(l.Logger), keyvals...)
}

func (l *kitLogger) warnLogger(keyvals []interface{}) log.Logger {
	return log.WithPrefix(level.Warn(l.Logger), keyvals...)
}

func (l *kitLogger) errorLogger(keyvals []interface{}) log.Logger {
	return log.WithPrefix(level.Error(l.Logger), keyvals...)
}
