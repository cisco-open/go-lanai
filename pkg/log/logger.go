package log

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// to avoid copy as possible, we reuse underlying array
// note, kv pairs's order should not matter
func buildLogEntry(name string, msg string, args []interface{}) []interface{} {
	return append(args, LogKeyName, name, LogKeyMessage, msg)
}

type CompositeKitLogger struct {
	delegates []log.Logger
}

func (c *CompositeKitLogger) addLogger(l log.Logger) {
	c.delegates = append(c.delegates, l)
}

func (c *CompositeKitLogger) Log(keyvals ...interface{}) error {
	for _, d := range c.delegates {
		_ = d.Log(keyvals...)
	}
	return nil
}

// logger implements Logger
type logger struct {
	log.Logger
}

func (l *logger) WithKV(keyvals...interface{}) Logger {
	return l.withKV(keyvals)
}

func (l *logger) Debugf(msg string, args ...interface{}) {

	l.debugLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *logger) Infof(msg string, args ...interface{}) {
	l.infoLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *logger) Warnf(msg string, args ...interface{}) {
	l.warnLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *logger) Errorf(msg string, args ...interface{}) {
	l.errorLogger([]interface{}{
		LogKeyMessage, fmt.Sprintf(msg, args...),
	}).Log()
}

func (l *logger) Debug(msg string, keyvals... interface{}) {
	l.debugLogger(keyvals).Log(LogKeyMessage, msg)
}

func (l *logger) Info(msg string, keyvals... interface{}) {
	l.infoLogger(keyvals).Log(LogKeyMessage, msg)
}

func (l *logger) Warn(msg string, keyvals... interface{}) {
	l.warnLogger(keyvals).Log(LogKeyMessage, msg)
}

func (l *logger) Error(msg string, keyvals... interface{}) {
	l.errorLogger(keyvals).Log(LogKeyMessage, msg)
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
