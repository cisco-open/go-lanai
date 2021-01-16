package log

import (
	"fmt"
	"github.com/go-kit/kit/log"
	kit_logurs "github.com/go-kit/kit/log/logrus"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

const levelDefault = "default"
const formatJson = "json"
const outputConsole = "console"

type kitLoggerFactory struct {
	rootLogLevel   LoggingLevel
	logLevels      map[string]LoggingLevel
	templateLogger log.Logger
	extractors     []FieldsExtractor
	registry       map[string]*kitLogger
}

func newKitLoggerFactory(properties *Properties) *kitLoggerFactory {
	rootLogLevel, ok := properties.Levels[levelDefault]
	if !ok {
		rootLogLevel = LevelInfo
	}
	template := buildTemplateLoggerFromConfig(properties)

	return &kitLoggerFactory{
		rootLogLevel: rootLogLevel,
		logLevels: properties.Levels,
		templateLogger: template,
		registry: make(map[string]*kitLogger),
	}
}

func (f *kitLoggerFactory) createLogger (name string) ContextualLogger{
	if l, ok := f.registry[name]; ok {
		return l
	}

	ll, ok := f.logLevels[name]
	if !ok {
		ll = f.rootLogLevel
	}

	l := newKitLogger(name, f.templateLogger, ll, f.extractors)
	f.registry[name] = l
	return l
}

func (f *kitLoggerFactory) addExtractors (extractors... FieldsExtractor) {
	f.extractors = append(f.extractors, extractors...)
}

func (f *kitLoggerFactory) setLevel (name string, logLevel LoggingLevel) {
	if logger, ok := f.registry[name]; ok {
		logger.setLevel(logLevel)
	}
}

func (f *kitLoggerFactory) refresh (properties *Properties) {
	rootLogLevel, ok := properties.Levels[levelDefault]
	if !ok {
		rootLogLevel = LevelInfo
	}
	template := buildTemplateLoggerFromConfig(properties)

	f.templateLogger = template
	f.rootLogLevel = rootLogLevel
	f.logLevels = properties.Levels

	for name, logger := range f.registry {
		ll, ok := f.logLevels[name]
		if !ok {
			ll = rootLogLevel
		}
		logger.templateLogger = f.templateLogger
		logger.setLevel(ll)
	}
}

func buildTemplateLoggerFromConfig(properties *Properties) log.Logger {
	composite := &CompositeKitLogger{}
	for output, format := range properties.Logger {
		switch output {
		case outputConsole:
			if format == formatJson {
				composite.addLogger(log.NewJSONLogger(log.NewSyncWriter(os.Stdout)))
			} else {
				composite.addLogger(makeTextLogger(log.NewSyncWriter(os.Stdout)))
			}
		default:
			composite.addLogger(makeTextLogger(log.NewSyncWriter(os.Stdout)))
		}
	}

	if len(composite.delegates) == 0 {
		composite.addLogger(makeTextLogger(log.NewSyncWriter(os.Stdout)))
	}
	return log.With(composite, "ts", log.DefaultTimestampUTC, "caller", log.Caller(7))
}

type CustomFormatter struct {
	*logrus.TextFormatter
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Level, _ = logrus.ParseLevel(entry.Data["level"].(fmt.Stringer).String())
	entry.Message = entry.Data[messageKey].(string)
	delete(entry.Data, "ts")
	delete(entry.Data, "level")
	delete(entry.Data, messageKey)
	return f.TextFormatter.Format(entry)
}

func makeTextLogger(w io.Writer) log.Logger{
	logrusLogger := logrus.New()
	logrusLogger.Out = w
	formatter := &CustomFormatter{
		&logrus.TextFormatter{},
	}
	//TODO: sort fields
	logrusLogger.SetFormatter(formatter)
	logger := kit_logurs.NewLogrusLogger(logrusLogger)
	return logger
}