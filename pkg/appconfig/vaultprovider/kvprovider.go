package vaultprovider

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
)

var logger = log.New("vaultprovider")

//Vault kv v1 differs with v2 API both in how the context path is constructed and how the response is parsed.
//https://www.vaultproject.io/docs/secrets/kv/kv-v1
type KeyValueConfigProvider struct {
	appconfig.ProviderMeta
	secretPath	string
	secretEngine KvSecretEngine
}



func (p *KeyValueConfigProvider) Name() string {
	return fmt.Sprintf("vault:%s", p.secretEngine.ContextPath(p.secretPath))
}

func (p *KeyValueConfigProvider) Load(ctx context.Context) (loadError error) {
	defer func(){
		if loadError != nil {
			p.Loaded = false
		} else {
			p.Loaded = true
		}
	}()

	p.Settings = make(map[string]interface{})

	// load keys from default context
	var defaultSettings map[string]interface{}

	defaultSettings, loadError = p.secretEngine.ListSecrets(ctx, p.secretPath)
	if loadError != nil {
		return loadError
	}

	unFlattenedSettings, loadError := appconfig.UnFlatten(defaultSettings)
	if loadError != nil {
		return loadError
	}

	p.Settings = unFlattenedSettings

	return nil
}

func NewVaultKvProvider(precedence int, secretPath string, secretEngine KvSecretEngine) *KeyValueConfigProvider {
	return &KeyValueConfigProvider{
		ProviderMeta: appconfig.ProviderMeta{Precedence: precedence},
		secretPath: secretPath,
		secretEngine: secretEngine,
	}
}