package log

import (
	"embed"
	"encoding/json"
	"github.com/ghodss/yaml"
	"io"
	"io/fs"
	"os"
	"path"
	"reflect"
	"strings"
)

// factory is created by init, and used to create new loggers.
var (
	factory       *zapLoggerFactory
	defaultConfig *Properties
)

// New is the intuitive starting point for any packages to use log package
// it will create a named logger if a logger with this name doesn't exist yet
func New(name string) ContextualLogger {
	return factory.createLogger(name)
}

func RegisterContextLogFields(extractors ...ContextValuers) {
	factory.addContextValuers(extractors...)
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

	//factory = newKitLoggerFactory(defaultConfig)
	factory = newZapLoggerFactory(defaultConfig)

	// a test run for dev
	//DebugShowcase()
}

func loadConfig(fs fs.FS, path string) (*Properties, error) {
	file, e := fs.Open(path)
	if e != nil {
		return nil, e
	}

	encoded, e := io.ReadAll(file)
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

	normalizeProperties(props)
	return props, nil
}

// normalizeProperties updates all KVs to lower case, which is consistent with appconfig binding
func normalizeProperties(props *Properties) {
	val := reflect.ValueOf(props).Elem()
	for i := val.Type().NumField() - 1; i >= 0; i-- {
		fv := val.Field(i)
		fv.Set(normalizeMapKeys(fv))
	}
}

// normalizeMapKeys updates all keys to lower case, which is consistent with appconfig binding
func normalizeMapKeys(mapValue reflect.Value) reflect.Value {
	typ := mapValue.Type()
	if typ.Kind() != reflect.Map || typ.Key().Kind() != reflect.String {
		return mapValue
	}

	ret := reflect.MakeMap(typ)
	iter := mapValue.MapRange()
	for iter.Next() {
		k := iter.Key()
		str := strings.ToLower(k.String())
		v := normalizeMapKeys(iter.Value())
		ret.SetMapIndex(reflect.ValueOf(str), v)
	}
	return ret
}
