package appconfig

import (
	"encoding/json"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"sort"
	"strings"
)

var ErrNotLoaded = errors.New("Configuration not loaded")
var ErrNotFound = errors.New("Missing required setting")

type Config struct {
	Providers     []Provider //such as yaml provider, commandline etc.
	settings      map[string]interface{}
	isLoaded 	  bool
}

type ConfigAccessor interface {
	Value(key string) (interface{}, error)
	Bind(target interface{}, prefix string) error
	Each(apply func(string, interface{}))
}

func NewConfig(providers ...Provider) *Config {
	return &Config{
		Providers:     providers,
		settings:      nil,
	}
}

func (c *Config) Load(force bool) (loadError error) {
	defer func() {
		if loadError != nil {
			c.isLoaded = false
		} else {
			c.isLoaded = true
		}
	}()

	//sort based on precedence
	sort.SliceStable(c.Providers, func(i, j int) bool { return c.Providers[i].GetPrecedence() > c.Providers[j].GetPrecedence() })

	// Load appconfig from each provider if it's not loaded yet, or if force reload.
	for _, provider := range c.Providers {
		if !provider.isLoaded() || force {
			error := provider.Load()

			if error != nil {
				return errors.Wrap(error, "Failed to load properties")
			}
		}
	}

	//TODO: resolve place holder

	merged := make(map[string]interface{})
	// merge data
	for _, provider := range c.Providers {
		error := mergo.Merge(&merged, provider.GetSettings())

		if error != nil {
			return errors.Wrap(error, "Failed to merge properties from property sources")
		}
	}

	c.settings = merged
	return nil
}

func (c *Config) Value(key string) (interface{}, error) {
	if !c.isLoaded {
		return "", ErrNotLoaded
	}

	targetKey := c.alias(key)

	nestedKeys := UnFlattenKey(targetKey)

	var tmp interface{} = c.settings
	for _, k := range nestedKeys {
		//TODO: case check
		tmp = (tmp.(map[string]interface{}))[k]
	}

	return tmp, nil
}

func (c *Config) Bind(target interface{}, prefix string) error {
	keys := strings.Split(prefix, ".")

	var source interface{} = c.settings

	//TODO: switch on type
	for i := 0; i < len(keys); i++ {
		source = source.(map[string] interface{})[keys[i]]
	}

	//TODO: error handling
	serialized, error := json.Marshal(source)

	if error == nil {
		error = json.Unmarshal(serialized, target)
	}


	return nil
}

func (c *Config) Each(apply func(string, interface{})) {
	VisitEach(c.settings, apply)
}

func (c *Config) alias(key string) string {
	// Return the actual target key name mapping through aliases
	return NormalizeKey(key)
}
