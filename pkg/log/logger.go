package log

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type kitLogger struct {
	swapLogger     *log.SwapLogger
	templateLogger log.Logger
	name           string
	extractors     []FieldsExtractor
}

func newKitLogger(name string, templateLogger log.Logger, logLevel LoggingLevel, extractors []FieldsExtractor) *kitLogger {
	s := &log.SwapLogger{}
	k := &kitLogger{
		name: name,
		swapLogger: s,
		templateLogger: templateLogger,
		extractors: extractors,
	}
	k.setLevel(logLevel)
	return k
}

func (k *kitLogger) Debug(msg string, args... interface{}) {
	_ = level.Debug(k.swapLogger).Log(buildLogEntry(k.name, msg, args...)...)
}
func (k *kitLogger) Info(msg string, args... interface{}) {
	_ = level.Info(k.swapLogger).Log(buildLogEntry(k.name, msg, args...)...)
}
func (k *kitLogger) Warn(msg string, args... interface{}) {
	_ = level.Warn(k.swapLogger).Log(buildLogEntry(k.name, msg, args...)...)
}
func (k *kitLogger) Error(msg string, args... interface{}) {
	_ = level.Error(k.swapLogger).Log(buildLogEntry(k.name, msg, args...)...)
}
func (k *kitLogger) WithContext(ctx context.Context) Logger {
	var ctxFields []interface{}
	if ctx != nil {
		for _, e := range k.extractors {
			m := e(ctx)
			for k, v := range m {
				ctxFields = append(ctxFields, k, v)
			}
		}
	}

	return &ContextLogger{
		ctxFields: ctxFields,
		swapLogger: k.swapLogger,
		name: k.name,
	}
}

func (k *kitLogger) setLevel(l LoggingLevel) {
	var opt level.Option
	switch l {
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
	logger := level.NewFilter(k.templateLogger, opt)
	k.swapLogger.Swap(logger)
}

func (k *kitLogger) getName() string {
	return k.name
}

func buildLogEntry(name string, msg string, args...interface{}) []interface{} {
	kvs := make([]interface{}, len(args)+4)
	kvs[0], kvs[1] = nameKey, name
	kvs[2], kvs[3] = messageKey, msg
	copy(kvs[4:], args)
	return kvs
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

type ContextLogger struct {
	ctxFields []interface{}
	swapLogger *log.SwapLogger
	name string
}

func (l *ContextLogger) Debug(msg string, args... interface{}) {
	_ = level.Debug(l.swapLogger).Log(buildLogEntry(l.name, msg, combineSlices(l.ctxFields, args)...)...)
}
func (l *ContextLogger) Info(msg string, args... interface{}) {
	_ = level.Info(l.swapLogger).Log(buildLogEntry(l.name, msg, combineSlices(l.ctxFields, args)...)...)
}
func (l *ContextLogger) Warn(msg string, args... interface{}) {
	_ = level.Warn(l.swapLogger).Log(buildLogEntry(l.name, msg, combineSlices(l.ctxFields, args)...)...)
}
func (l *ContextLogger) Error(msg string, args... interface{}) {
	_ = level.Error(l.swapLogger).Log(buildLogEntry(l.name, msg, combineSlices(l.ctxFields, args)...)...)
}

func combineSlices(first []interface{}, second []interface{}) []interface{} {
	mergedArgs := make([]interface{}, len(first) + len(second))
	copy(mergedArgs[:], first)
	copy(mergedArgs[len(first):], second)
	return mergedArgs
}