package consulprovider

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig"
	"cto-github.cisco.com/livdu/jupiter/pkg/consul"
	"fmt"
)

const (
	ConfigRootConsulConfigProvider = "spring.cloud.consul.appconfig" //TODO: name should be changed
	ConfigKeyAppName               = "spring.application.name"
)

type ConsulConfigProperties struct {
	Enabled        bool   `json:enabled`
	Prefix         string `json:prefix`
	DefaultContext string `json:"default-context`
}

type ConfigProvider struct {
	appconfig.ProviderMeta
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
	unFlattenedSettings, _ := appconfig.UnFlatten(defaultSettings)

	f.Settings = unFlattenedSettings

	f.Valid = true
}

func NewConsulProvider(description string, precedence int, contextPath string, conn *consul.Connection) *ConfigProvider {
	return &ConfigProvider{
			ProviderMeta: appconfig.ProviderMeta{Description: description, Precedence: precedence},
			contextPath:  contextPath, //fmt.Sprintf("%s/%s", f.sourceConfig.Prefix, f.contextPath)
			connection:   conn,
		}
}
