package vaultprovider

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"fmt"
)

type GenericConfigProvider struct {
	appconfig.ProviderMeta
	secretEngine *vault.GenericSecretEngine
	contextPath  string
}

func (configProvider *GenericConfigProvider) Name() string {
	return fmt.Sprintf("vault:%s", configProvider.contextPath)
}

func (configProvider *GenericConfigProvider) Load() (loadError error) {
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

	defaultSettings, loadError = configProvider.secretEngine.ListSecrets(bootstrap.EagerGetApplicationContext(), configProvider.contextPath)
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

func NewVaultGenericProvider(precedence int, contextPath string, conn *vault.Connection) *GenericConfigProvider {
	return &GenericConfigProvider{
		ProviderMeta: appconfig.ProviderMeta{Precedence: precedence},
		contextPath:  contextPath,
		secretEngine: conn.GenericSecretEngine(),
	}
}