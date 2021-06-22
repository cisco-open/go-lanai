package log

import (
	"embed"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"io/fs"
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

//go:embed defaults-log.yml
var defaultConfigFS embed.FS

// Since log package cannot depend on other packages in case those package want to use log,
// we have to duplicate the code for reading profile here.
func init() {
	fullPath := path.Join("configs", "log.yml")
	var err error
	if stat, e := os.Stat(fullPath); e == nil && !stat.IsDir() {
		// log.yml is available, try use it
		defaultConfig, err = loadConfig(os.DirFS("."), fullPath)
	}

	if err != nil || defaultConfig == nil {
		// log.yml is not available, uses embedded defaults
		defaultConfig, err = loadConfig(defaultConfigFS, "defaults-log.yml")
	}

	if err != nil || defaultConfig == nil {
		defaultConfig = newProperties()
	}

	factory = newKitLoggerFactory(defaultConfig)

	// a test run for dev
	//DebugShowcase()
}

func loadConfig(fs fs.FS, path string) (*Properties, error) {
	file, e := fs.Open(path)
	if e != nil {
		return nil, e
	}

	encoded, e := ioutil.ReadAll(file)
	if e != nil {
		return nil, e
	}

	encodedJson, e := yaml.YAMLToJSON(encoded)
	if e != nil {
		return nil, e
	}

	props := newProperties()
	if e := json.Unmarshal(encodedJson, props); e != nil {
		return nil, e
	}
	return props, nil
}