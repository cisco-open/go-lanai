package log

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"io/ioutil"
	"os"
	"path"
)

// factory is created by init, and used to create new loggers.
var (
	factory       loggerFactory
	defaultConfig *Properties
)

// New is the intuitive starting point for any packages to use log package
// it will create a named logger if a logger with this name doesn't exist yet
func New(name string) ContextualLogger {
	return factory.createLogger(name)
}

func RegisterContextLogFields(extractors...ContextValuers) {
	factory.addContextValuers(extractors...)
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

// Since log package cannot depend on other packages in case those package want to use log,
// we have to duplicate the code for reading profile here.

func init() {
	defaultConfig = newProperties()

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

	// a test run for dev
	internal.DebugShowcase()
}