package log

import (
	"context"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"io/ioutil"
	"os"
	"path"
)

// factory is created by init, and used to create new loggers.
var factory LoggerFactory
var defaultConfig *Properties

type LoggingLevel string

const (
	LevelDebug LoggingLevel = "debug"
	LevelInfo LoggingLevel = "info"
	LevelWarn LoggingLevel = "warn"
	LevelError LoggingLevel = "error"
)

type FieldsExtractor func (ctx context.Context) map[string]interface{}

type ContextualLogger interface {
	Logger
	Contextual
}

type Contextual interface {
	WithContext(ctx context.Context) Logger
}

type Logger interface {
	Debug(msg string, args... interface{})
	Info(msg string, args... interface{})
	Warn(msg string, args... interface{})
	Error(msg string, args... interface{})
}

type LoggerFactory interface {
	 createLogger (name string) ContextualLogger
	 addExtractors (extractors...FieldsExtractor)
	 setLevel (name string, logLevel LoggingLevel)
	 refresh (properties *Properties)
}

func RegisterContextLogFields(extractors...FieldsExtractor) {
	factory.addExtractors(extractors...)
}

//Will create a new logger if a logger with this name doesn't exist yet
func GetNamedLogger(name string) ContextualLogger {
	logger := factory.createLogger(name)
	return logger
}

func SetLevel(name string, logLevel LoggingLevel) {
	factory.setLevel(name, logLevel)
}

func UpdateLoggingConfiguration(properties *Properties) error {
	mergedProperties := &Properties{}
	mergeOption := func(mergoConfig *mergo.Config) {
		mergoConfig.Overwrite = true
	}
	err := mergo.Merge(mergedProperties, defaultConfig, mergeOption)
	if err != nil {
		return err
	}
	err = mergo.Merge(mergedProperties, properties, mergeOption)
	if err != nil {
		return err
	}
	factory.refresh(mergedProperties)
	return err
}

//Since log package cannot depend on other packages in case those package want to use log,
//we have to duplicate the code for reading profile here.

func init() {
	defaultConfig = &Properties{
		Levels: map[string]LoggingLevel{"default": LevelInfo},
	}

	fullPath := path.Join("configs", "log.yml")
	info, err := os.Stat(fullPath)
	if !os.IsNotExist(err) && !info.IsDir() {
		file, err := os.Open(fullPath)
		if err == nil {
			encoded, err := ioutil.ReadAll(file)
			if err == nil {
				encodedJson, err := yaml.YAMLToJSON(encoded)
				if err == nil {
					err = json.Unmarshal(encodedJson, defaultConfig)
				}
			}
		}
	}
	factory = newKitLoggerFactory(defaultConfig)
}