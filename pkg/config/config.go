package config

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"sort"
	"strings"
)

var ErrNotLoaded = errors.New("Configuration not loaded")
var ErrNotFound = errors.New("Missing required setting")

type Config struct {
	Providers     []Provider //such as yaml provider, commandline etc.
	settings      map[string]interface{}
}

func NewConfig(providers ...Provider) *Config {
	return &Config{
		Providers:     providers,
		settings:      nil,
	}
}

func (c *Config) AddProvider(provider Provider) {
	c.Providers = append(c.Providers, provider)
}

//TODO: is this function necessary? or correct? just because settings is not nil doesn't necessarily mean loading is complete
func (c *Config) Loaded() bool {
	return c.settings != nil
}

func (c *Config) Load(force bool) error {
	//sort based on precedence
	sort.SliceStable(c.Providers, func(i, j int) bool { return c.Providers[i].GetPrecedence() > c.Providers[j].GetPrecedence() })

	// Load config from each provider if it's not loaded yet, or if force reload.
	for _, provider := range c.Providers {
		if !provider.isValid() || force {
			provider.Load()
		}
	}

	settings := make(map[string]interface{})
	// merge data
	for _, provider := range c.Providers {
		for k, v := range provider.GetSettings() {
			settings[k] = v
		}
	}

	// Resolve variables in the config
	if err := c.resolve(settings); err != nil {
		return errors.Wrap(err, "Failed to resolve variables")
	}

	c.settings = settings
	return nil
}

//TODO: review the implementation
func (c *Config) resolveValueHelper(nested map[string]bool,
	resolved, settings map[string]interface{}, value string) (string, error) {
	variableRegex, _ := regexp.Compile(`\${([\w._\-]+)(:(.*))?}`)
	if !strings.Contains(value, "${") {
		return value, nil
	}

	var refStrings []string
	depth := 0
	refString := ""
	for i, c := range value {
		if value[i] == '$' && i+1 < len(value) && value[i+1] == '{' {
			depth++
		}
		if depth > 0 {
			refString += string(c)
		}
		if value[i] == '}' && depth > 0 {
			depth--
		}
		if depth == 0 && len(refString) > 0 {
			refStrings = append(refStrings, refString)
			refString = ""
		}
	}

	if depth != 0 {
		fmt.Println("Malformed string %s", value)
		// return raw string
		return value, nil
	}

	for _, rs := range refStrings {
		match := variableRegex.FindStringSubmatch(rs)
		if len(match) > 0 {
			//ignore case
			refName := c.alias(match[1])
			refValue := ""
			//check if variable is already in the stack
			if _, ok := nested[refName]; ok {
				return "", errors.Errorf("Circular variable reference detected: '%s' already in %v",
					refName, nested)
			} else {
				nested[refName] = true
			}
			// check lasted resolved values first
			if rv, ok := resolved[refName]; ok {
				refValue = rv.(string)
			} else if sv, ok := settings[refName]; ok {
				refValue = sv.(string)
			}
			// resolve the references in the value.
			refValue, err := c.resolveValueHelper(nested, resolved, settings, refValue)
			if err != nil {
				return "", err
			}
			if len(refValue) > 0 {
				resolved[refName] = refValue
			} else if len(match) >= 4 && len(match[3]) > 0 {
				// if not able to resolve the value and default value is set
				refValue, err = c.resolveValueHelper(nested, resolved, settings, match[3])
			}
			if err != nil {
				return "", err
			}
			//remove resolved variable from nested map.
			delete(nested, refName)
			value = strings.Replace(value, rs, refValue, -1)
		}
	}
	return value, nil
}

// Expand all references to ${variables} inside value
func (c *Config) resolveValue(resolved, settings map[string]interface{}, value string) string {
	nested := make(map[string]bool)
	value, err := c.resolveValueHelper(nested, resolved, settings, value)
	if err != nil {
		fmt.Println("Not able to resolve value!")
		return ""
	}
	return value
}

// Expand all references to ${variables}
func (c *Config) resolve(settings map[string]interface{}) error {
	resolved := make(map[string]interface{})

	for k, v := range settings {
		resolved[k] = c.resolveValue(resolved, settings, v.(string))
	}

	for k, v := range resolved {
		settings[k] = v
	}

	return nil
}

func (c *Config) Value(key string) (interface{}, error) {
	if !c.Loaded() {
		return "", ErrNotLoaded
	}

	targetKey := c.alias(key)
	if val, ok := c.settings[targetKey]; !ok {
		return "", errors.Wrap(ErrNotFound, key)
	} else {
		return val, nil
	}
}

func (c *Config) Populate(target interface{}, prefix string) error {
	// Wrap the properties map in a partial config
	stringSettings := make(map[string]string)
	for k, v := range c.settings {
		stringSettings[k] = fmt.Sprintf("%v", v)
	}

	partialConfig := NewPartialConfig(stringSettings, c)
	// Filter by prefix
	partialConfig = partialConfig.FilterStripPrefix(NormalizeKey(prefix))

	// Populate the object from the properties map
	return partialConfig.Populate(target)
}

func (c *Config) alias(key string) string {
	// Return the actual target key name mapping through aliases
	return NormalizeKey(key)
}

func (c *Config) Each(target func(string, interface{})) {
	for name, value := range c.settings {
		target(name, value)
	}
}