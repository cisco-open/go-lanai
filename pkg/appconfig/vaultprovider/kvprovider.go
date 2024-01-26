// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package vaultprovider

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
)

var logger = log.New("Config.Vault")

// KeyValueConfigProvider
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

	logger.WithContext(ctx).Infof("Retrieved %d secrets from vault path: %s", len(defaultSettings), p.secretEngine.ContextPath(p.secretPath))
	return nil
}

func NewVaultKvProvider(precedence int, secretPath string, secretEngine KvSecretEngine) *KeyValueConfigProvider {
	return &KeyValueConfigProvider{
		ProviderMeta: appconfig.ProviderMeta{Precedence: precedence},
		secretPath: secretPath,
		secretEngine: secretEngine,
	}
}