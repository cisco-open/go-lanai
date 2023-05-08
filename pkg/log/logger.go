package log

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
// logger implements Logger
type logger struct {
	log.Logger
}

func (l *logger) WithKV(keyvals...interface{}) Logger {
	return l.withKV(keyvals)
}

func (l *logger) WithLevel(lvl LoggingLevel) Logger {
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

	return &logger{
		Logger: leveled,
	}
}

func (l *logger) WithCaller(caller interface{}) Logger {
	return &logger{
		Logger: log.WithSuffix(l.Logger, LogKeyCaller, caller),
	}
}

func (l *logger) Debugf(msg string, args ...interface{}) {
	_ = l.debugLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *logger) Infof(msg string, args ...interface{}) {
	_ = l.infoLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *logger) Warnf(msg string, args ...interface{}) {
	_ = l.warnLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *logger) Errorf(msg string, args ...interface{}) {
	_ = l.errorLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *logger) Debug(msg string, keyvals... interface{}) {
	_ = l.debugLogger(keyvals).Log(LogKeyMessage, msg)
}

func (l *logger) Info(msg string, keyvals... interface{}) {
	_ = l.infoLogger(keyvals).Log(LogKeyMessage, msg)
}

func (l *logger) Warn(msg string, keyvals... interface{}) {
	_ = l.warnLogger(keyvals).Log(LogKeyMessage, msg)
}

func (l *logger) Error(msg string, keyvals... interface{}) {
	_ = l.errorLogger(keyvals).Log(LogKeyMessage, msg)
}


func (l *logger) Print(args ...interface{})  {
	_ = l.Log(LogKeyMessage, fmt.Sprint(args...))
}

func (l *logger) Printf(format string, args ...interface{}) {
	_ = l.Log(LogKeyMessage, fmt.Sprintf(format, args...))
}

func (l *logger) Println(args ...interface{}) {
	_ = l.Log(LogKeyMessage, fmt.Sprintln(args...))
}

func (l *logger) withKV(keyvals []interface{}) Logger {
	return &logger{
		Logger: log.WithPrefix(l.Logger, keyvals...),
	}
}

func (l *logger) debugLogger(keyvals []interface{}) log.Logger {
	return log.WithPrefix(level.Debug(l.Logger), keyvals...)
}

func (l *logger) infoLogger(keyvals []interface{}) log.Logger {
	return log.WithPrefix(level.Info(l.Logger), keyvals...)
}

func (l *logger) warnLogger(keyvals []interface{}) log.Logger {
	return log.WithPrefix(level.Warn(l.Logger), keyvals...)
}

func (l *logger) errorLogger(keyvals []interface{}) log.Logger {
	return log.WithPrefix(level.Error(l.Logger), keyvals...)
}
