package log

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var zapEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        LogKeyTimestamp,
	LevelKey:       LogKeyLevel,
	NameKey:        LogKeyName,
	CallerKey:      LogKeyCaller,
	FunctionKey:    zapcore.OmitKey,
	MessageKey:     LogKeyMessage,
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime: func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		// RFC339 with milliseconds
		encoder.AppendString(time.UTC().Format(`2006-01-02T15:04:05.999Z07:00`))
	},
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller: func(ec zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
		if !ec.Defined || len(ec.File) == 0 {
			return
		}
		idx := strings.LastIndexByte(ec.File, '/')
		encoder.AppendString(ec.File[idx+1:] + ":" + strconv.Itoa(ec.Line))
	},
}

type zapCoreCreator func(level zap.AtomicLevel) zapcore.Core

type zapLoggerFactory struct {
	rootLogLevel     LoggingLevel
	logLevels        map[string]LoggingLevel
	coreCreator      zapCoreCreator
	properties       *Properties
	effectiveValuers ContextValuers
	extraValuers     ContextValuers
	registry         map[string]*configurableZapLogger
}

func newZapLoggerFactory(properties *Properties) *zapLoggerFactory {
	rootLogLevel, ok := properties.Levels[keyLevelDefault]
	if !ok {
		rootLogLevel = LevelInfo
	}

	var e error
	f := &zapLoggerFactory{
		rootLogLevel:     rootLogLevel,
		logLevels:        properties.Levels,
		properties:       properties,
		registry:         map[string]*configurableZapLogger{},
		extraValuers:     ContextValuers{},
		effectiveValuers: ContextValuers{},
	}
	f.effectiveValuers = f.buildContextValuer(properties)
	if f.coreCreator, e = f.buildZapCoreCreator(properties); e != nil {
		panic(e)
	}
	return f
}

func (f *zapLoggerFactory) createLogger(name string) ContextualLogger {
	key := loggerKey(name)
	if l, ok := f.registry[key]; ok {
		return l
	}

	ll := f.resolveEffectiveLevel(key)
	leveler := zap.NewAtomicLevel()
	l := newConfigurableZapLogger(name, f.coreCreator(leveler), ll, leveler, f.effectiveValuers)
	f.registry[key] = l
	return l
}

func (f *zapLoggerFactory) addContextValuers(valuers ...ContextValuers) {
	for _, item := range valuers {
		for k, v := range item {
			f.effectiveValuers[k] = v
			f.extraValuers[k] = v
		}
	}
}

func (f *zapLoggerFactory) setRootLevel(logLevel LoggingLevel) (affected int) {
	f.rootLogLevel = logLevel
	for k, l := range f.registry {
		effective := f.resolveEffectiveLevel(k)
		l.setMinLevel(effective)
		affected++
	}
	return
}

func (f *zapLoggerFactory) setLevel(prefix string, logLevel *LoggingLevel) (affected int) {
	key := loggerKey(prefix)
	if (key == "" || key == keyLevelDefault || key == loggerKey(nameLevelDefault)) && logLevel != nil {
		return f.setRootLevel(*logLevel)
	}

	if logLevel == nil {
		// unset
		if _, ok := f.logLevels[key]; ok {
			delete(f.logLevels, key)
		}
	} else {
		// set
		f.logLevels[key] = *logLevel
	}

	// set effective level to all affected loggers
	withDot := key + keySeparator
	for k, l := range f.registry {
		if k != key && !strings.HasPrefix(k, withDot) {
			continue
		}
		effective := f.resolveEffectiveLevel(k)
		l.setMinLevel(effective)
		affected++
	}
	return
}

func (f *zapLoggerFactory) refresh(properties *Properties) {
	rootLogLevel, ok := properties.Levels[keyLevelDefault]
	if !ok {
		rootLogLevel = LevelInfo
	}

	f.rootLogLevel = rootLogLevel
	f.logLevels = properties.Levels
	f.effectiveValuers = buildContextValuerFromConfig(properties)

	// merge valuers, note: we don't delete extra valuers during refresh
	for k, v := range f.extraValuers {
		f.effectiveValuers[k] = v
	}

	for key, l := range f.registry {
		ll := f.resolveEffectiveLevel(key)
		l.valuers = f.effectiveValuers
		l.setMinLevel(ll)
	}
}

func (f *zapLoggerFactory) resolveEffectiveLevel(key string) LoggingLevel {
	prefix := key
	for i := len(key); i > 0; i = strings.LastIndex(prefix, keySeparator) {
		prefix = key[0:i]
		if ll, ok := f.logLevels[prefix]; ok {
			return ll
		}
	}
	return f.rootLogLevel
}

func (f *zapLoggerFactory) loggerKey(name string) string {
	return utils.CamelToSnakeCase(name)
}

func (f *zapLoggerFactory) buildContextValuer(properties *Properties) ContextValuers {
	valuers := ContextValuers{}
	// k is context-key, v is log-key
	for k, v := range properties.Mappings {
		valuers[v] = func(ctx context.Context) interface{} {
			return ctx.Value(k)
		}
	}

	return valuers
}

func (f *zapLoggerFactory) buildZapCoreCreator(properties *Properties) (zapCoreCreator, error) {
	if len(properties.Loggers) == 0 {
		properties.Loggers["default"] = &LoggerProperties{
			Type:      TypeConsole,
			Format:    FormatText,
			Template:  defaultTemplate,
			FixedKeys: defaultFixedFields.Values(),
		}
	}
	encoders := make([]zapcore.Encoder, len(properties.Loggers))
	syncers := make([]zapcore.WriteSyncer, len(properties.Loggers))
	var i int
	for _, loggerProps := range properties.Loggers {
		var e error
		if syncers[i], e = f.newZapWriteSyncer(loggerProps); e != nil {
			return nil, e
		}
		if encoders[i], e = f.newZapEncoder(loggerProps, syncers[i].(internal.TerminalAware).IsTerminal()); e != nil {
			return nil, e
		}
		i++
	}
	return func(level zap.AtomicLevel) zapcore.Core {
		var core zapcore.Core
		var isTerm bool
		switch len(encoders) {
		case 0:
			// not possible
			return zapcore.NewNopCore()
		case 1:
			core = zapcore.NewCore(encoders[0], syncers[0], level)
			isTerm = syncers[0].(internal.TerminalAware).IsTerminal()
		default:
			cores := make([]zapcore.Core, len(encoders))
			for i := range encoders {
				cores[i] = zapcore.NewCore(encoders[i], syncers[i], level)
				isTerm = isTerm && syncers[i].(internal.TerminalAware).IsTerminal()
			}
			core = zapcore.NewTee(cores...)
		}
		if isTerm {
			return internal.ZapTerminalCore{Core: core}
		}
		return core
	}, nil
}

func (f *zapLoggerFactory) newZapEncoder(props *LoggerProperties, isTerm bool) (zapcore.Encoder, error) {
	switch props.Format {
	case FormatText:
		fixedFields := defaultFixedFields.Copy().Add(props.FixedKeys...)
		formatter := internal.NewTemplatedFormatter(props.Template, fixedFields, isTerm)
		return internal.NewZapFormattedEncoder(zapEncoderConfig, formatter, isTerm), nil
	case FormatJson:
		return zapcore.NewJSONEncoder(zapEncoderConfig), nil
	}
	return nil, fmt.Errorf("unsupported logger format: %v", props.Format)
}

func (f *zapLoggerFactory) newZapWriteSyncer(props *LoggerProperties) (zapcore.WriteSyncer, error) {
	switch props.Type {
	case TypeConsole:
		return internal.NewZapWriterWrapper(os.Stdout), nil
	case TypeFile:
		file, e := openOrCreateFile(props.Location)
		if e != nil {
			return nil, e
		}
		return internal.NewZapWriterWrapper(file), nil
	case TypeHttp:
		fallthrough
	case TypeMQ:
		fallthrough
	default:
		return nil, fmt.Errorf("unsupported logger type: %v", props.Type)
	}
}

func (f *zapLoggerFactory) openOrCreateFile(location string) (*os.File, error) {
	if location == "" {
		return nil, fmt.Errorf("location is missing for file logger")
	}
	dir := filepath.Dir(location)
	if e := os.MkdirAll(dir, 0744); e != nil {
		return nil, e
	}
	return os.OpenFile(location, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
}
