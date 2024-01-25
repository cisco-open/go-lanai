package log

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
	"fmt"
	"github.com/go-kit/kit/log"
	"io"
	"os"
	"strings"
)

const (
	keyLevelDefault  = "default"
	keySeparator     = "."
	nameLevelDefault = "ROOT"
)

type kitLoggerFactory struct {
	rootLogLevel     LoggingLevel
	logLevels        map[string]LoggingLevel
	templateLogger   log.Logger
	effectiveValuers ContextValuers
	extraValuers     ContextValuers
	registry         map[string]*configurableKitLogger
}

func newKitLoggerFactory(properties *Properties) *kitLoggerFactory {
	rootLogLevel, ok := properties.Levels[keyLevelDefault]
	if !ok {
		rootLogLevel = LevelInfo
	}

	return &kitLoggerFactory{
		rootLogLevel:     rootLogLevel,
		logLevels:        convertLevelsNameToKey(properties.Levels),
		templateLogger:   buildTemplateLoggerFromConfig(properties),
		registry:         map[string]*configurableKitLogger{},
		extraValuers:     ContextValuers{},
		effectiveValuers: buildContextValuerFromConfig(properties),
	}
}

func (f *kitLoggerFactory) createLogger(name string) ContextualLogger {
	key := loggerKey(name)
	if l, ok := f.registry[key]; ok {
		return l
	}

	ll := f.resolveEffectiveLevel(key)
	l := newConfigurableLogger(name, f.templateLogger, ll, f.effectiveValuers)
	f.registry[key] = l
	return l
}

func (f *kitLoggerFactory) addContextValuers(valuers ...ContextValuers) {
	for _, item := range valuers {
		for k, v := range item {
			f.effectiveValuers[k] = v
			f.extraValuers[k] = v
		}
	}
}

func (f *kitLoggerFactory) setRootLevel(logLevel LoggingLevel) (affected int) {
	f.rootLogLevel = logLevel
	for k, l := range f.registry {
		effective := f.resolveEffectiveLevel(k)
		l.setLevel(effective)
		affected++
	}
	return
}

func (f *kitLoggerFactory) setLevel(prefix string, logLevel *LoggingLevel) (affected int) {
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
		l.setLevel(effective)
		affected++
	}
	return
}

func (f *kitLoggerFactory) refresh(properties *Properties) error  {
	rootLogLevel, ok := properties.Levels[keyLevelDefault]
	if !ok {
		rootLogLevel = LevelInfo
	}

	f.templateLogger = buildTemplateLoggerFromConfig(properties)
	f.rootLogLevel = rootLogLevel
	f.logLevels = convertLevelsNameToKey(properties.Levels)
	f.effectiveValuers = buildContextValuerFromConfig(properties)

	// merge valuers, note: we don't delete extra valuers during refresh
	for k, v := range f.extraValuers {
		f.effectiveValuers[k] = v
	}

	for key, l := range f.registry {
		ll := f.resolveEffectiveLevel(key)
		l.template = f.templateLogger
		l.valuers = f.effectiveValuers
		l.setLevel(ll)
	}
	return nil
}

func (f *kitLoggerFactory) resolveEffectiveLevel(key string) LoggingLevel {
	prefix := key
	for i := len(key); i > 0; i = strings.LastIndex(prefix, keySeparator) {
		prefix = key[0:i]
		if ll, ok := f.logLevels[prefix]; ok {
			return ll
		}
	}
	return f.rootLogLevel
}

func buildContextValuerFromConfig(properties *Properties) ContextValuers {
	valuers := ContextValuers{}
	// k is context-key, v is log-key
	for k, v := range properties.Mappings {
		valuers[v] = func(ctx context.Context) interface{} {
			return ctx.Value(k)
		}
	}
	return valuers
}

func buildTemplateLoggerFromConfig(properties *Properties) log.Logger {
	composite := &compositeKitLogger{}
	for _, loggerProps := range properties.Loggers {
		logger, e := newKitLogger(loggerProps)
		if e != nil {
			panic(e)
		}
		composite.addLogger(logger)
	}

	logger := log.Logger(composite)
	switch len(composite.delegates) {
	case 0:
		defaultProps := &LoggerProperties{
			Type:      TypeConsole,
			Format:    FormatText,
			Template:  defaultTemplate,
			FixedKeys: defaultFixedFields.Values(),
		}
		logger, _ = newKitLogger(defaultProps)
	case 1:
		logger = composite.delegates[0]
	}
	return logger
}

func newKitLogger(props *LoggerProperties) (log.Logger, error) {
	switch props.Type {
	case TypeConsole:
		return newKitLoggerWithWriter(log.NewSyncWriter(os.Stdout), props)
	case TypeFile:
		f, e := openOrCreateFile(props.Location)
		if e != nil {
			return nil, e
		}
		return newKitLoggerWithWriter(log.NewSyncWriter(f), props)
	case TypeHttp:
	case TypeMQ:
	default:
	}
	return nil, fmt.Errorf("unsupported logger type: %v", props.Type)
}

func newKitLoggerWithWriter(w io.Writer, props *LoggerProperties) (log.Logger, error) {
	switch props.Format {
	case FormatText:
		fixedFields := defaultFixedFields.Copy().Add(props.FixedKeys...)
		formatter := internal.NewTemplatedFormatter(props.Template, fixedFields, internal.IsTerminal(w))
		return internal.NewKitTextLoggerAdapter(w, formatter), nil
	case FormatJson:
		return log.NewJSONLogger(w), nil
	}
	return nil, fmt.Errorf("unsupported logger format: %v", props.Format)
}

