package vaultprovider

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"fmt"
)

const (
	KvConfigPrefix = "cloud.vault.kv"
)

type KvConfigProperties struct {
	Enabled     string `json:"enabled"`
	Backend          string `json:"backend"`
	DefaultContext   string `json:"default-context"`
	ProfileSeparator string `json:"profile-separator"`
	ApplicationName  string `json:"application-name"`
}

type ConfigProvider struct {
	appconfig.ProviderMeta
	connection *vault.Connection
	contextPath string
}

func (configProvider *ConfigProvider) Name() string {
	return fmt.Sprintf("vault:%s", configProvider.contextPath)
}

func (configProvider *ConfigProvider) Load() (loadError error) {
	defer func(){
		if loadError != nil {
			configProvider.Loaded = false
		} else {
			configProvider.Loaded = true
		}
	}()

	configProvider.Settings = make(map[string]interface{})

	// load keys from default context
	var defaultSettings map[string]interface{}

	defaultSettings, loadError = configProvider.connection.ListSecrets(bootstrap.EagerGetApplicationContext(), configProvider.contextPath)
	if loadError != nil {
		return loadError
	}

	unFlattenedSettings, loadError := appconfig.UnFlatten(defaultSettings)
	if loadError != nil {
		return loadError
	}

	configProvider.Settings = unFlattenedSettings

	return nil
}

func NewVaultProvider(precedence int, contextPath string, conn *vault.Connection) *ConfigProvider {
	return &ConfigProvider{
		ProviderMeta: appconfig.ProviderMeta{Precedence: precedence},
		contextPath:  contextPath,
		connection:   conn,
	}
}