package log

import (
	"github.com/imdario/mergo"
	"strings"
)

// LevelConfig is a read-only carrier struct that stores LoggingLevel configuration of each logger
type LevelConfig struct {
	Name            string
	EffectiveLevel  *LoggingLevel
	ConfiguredLevel *LoggingLevel
}

// SetLevel set/unset logging level of all loggers with given prefix
// function returns actual number of affected loggers
func SetLevel(prefix string, logLevel *LoggingLevel) int {
	return factory.setLevel(prefix, logLevel)
}

// Levels logger level configuration, the returned map's key is the lower case of logger's name
func Levels(prefix string) (ret map[string]*LevelConfig) {
	ret = map[string]*LevelConfig{}
	prefixKey := loggerKey(prefix)

	// collect level config names
	for k, v := range factory.registry {
		if !strings.HasPrefix(k, prefixKey) {
			continue
		}
		p := v.name
		for i := len(v.name); i > 0; i = strings.LastIndex(p, keySeparator) {
			p = v.name[0 : i]
			ret[loggerKey(p)] = &LevelConfig{Name: p}
		}
	}
	// populate result
	for k, v := range ret {
		if l, ok := factory.registry[k]; ok {
			v.EffectiveLevel = &l.lvl
		} else {
			lvl := factory.resolveEffectiveLevel(k)
			v.EffectiveLevel = &lvl
		}
		if ll, ok := factory.logLevels[k]; ok {
			v.ConfiguredLevel = &ll
		}
	}
	if prefix == "" {
		ret[keyLevelDefault] = &LevelConfig{
			Name:            nameLevelDefault,
			ConfiguredLevel: &factory.rootLogLevel,
		}
	}
	return
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
