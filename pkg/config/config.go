package config

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

//TODO: change this to functional configuration
func (c *Config) Load(force bool) error {
	//sort based on precedence
	sort.SliceStable(c.Providers, func(i, j int) bool { return c.Providers[i].GetPrecedence() > c.Providers[j].GetPrecedence() })

	// Load config from each provider if it's not loaded yet, or if force reload.
	for _, provider := range c.Providers {
		if !provider.isValid() || force {
			provider.Load()
		}
	}

	merged := make(map[string]interface{})
	// merge data
	for _, provider := range c.Providers {
		error := mergo.Merge(&merged, provider.GetSettings())

		if error != nil {
			return error
		}
	}

	c.settings = merged
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

func (c *Config) alias(key string) string {
	// Return the actual target key name mapping through aliases
	return NormalizeKey(key)
}

//TODO: this needs to become recursive - similar to the flat method
func (c *Config) Each(apply func(string, interface{})) {
	VisitEach(c.settings, apply)
}

