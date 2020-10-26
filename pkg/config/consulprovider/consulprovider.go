package consulprovider

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/config"
	"cto-github.cisco.com/livdu/jupiter/pkg/consul"
	"fmt"
)

const (
	ConfigRootConsulConfigProvider = "spring.cloud.consul.config"
	ConfigKeyAppName               = "spring.application.name"
)

type ConfigProviderConfig struct {
	Enabled        bool   `config:"default=false"`
	Prefix         string `config:"default=userviceconfiguration"`
	DefaultContext string `config:"default=defaultapplication"`
}

type ConfigProvider struct {
	config.ProviderMeta
	contextPath  string
	connection   *consul.Connection
}

func (f *ConfigProvider) Load() {
	f.Valid = false

	f.Settings = make(map[string]interface{})

	// load keys from default context
	fmt.Printf("Loading configuration from consul (%s): %s)", f.connection.Host(), f.contextPath)
	var defaultSettings map[string]interface{}

	//TODO: error handling
	defaultSettings, _ = f.connection.ListKeyValuePairs(f.contextPath)

	for k, v := range defaultSettings {
		f.Settings[config.NormalizeKey(k)] = v
	}

	f.Valid = true
}

func NewConsulProvider(description string, precedence int, contextPath string, conn *consul.Connection) *ConfigProvider {
	return &ConfigProvider{
			ProviderMeta: config.ProviderMeta{Description: description, Precedence: precedence},
			contextPath:  contextPath, //fmt.Sprintf("%s/%s", f.sourceConfig.Prefix, f.contextPath)
			connection:   conn,
		}
}
