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

package consulprovider

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"fmt"
)

const (
	ConsulConfigPrefix = "cloud.consul.config"
	ConfigKeyAppName   = "application.name"
)

type ConsulConfigProperties struct {
	Enabled        bool   `json:"enabled"`
	Prefix         string `json:"prefix"`
	DefaultContext string `json:"default-context"`
	ProfileSeparator string `json:"profile-separator"`
}

type ConfigProvider struct {
	appconfig.ProviderMeta
	contextPath  string
	connection   *consul.Connection
}

func (configProvider *ConfigProvider) Name() string {
	return fmt.Sprintf("consul:%s", configProvider.contextPath)
}

func (configProvider *ConfigProvider) Load(ctx context.Context) (loadError error) {
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

	defaultSettings, loadError = configProvider.connection.ListKeyValuePairs(
		ctx,
		configProvider.contextPath)
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

func NewConsulProvider(precedence int, contextPath string, conn *consul.Connection) *ConfigProvider {
	return &ConfigProvider{
			ProviderMeta: appconfig.ProviderMeta{Precedence: precedence},
			contextPath:  contextPath, //fmt.Sprintf("%s/%s", f.sourceConfig.Prefix, f.contextPath)
			connection:   conn,
		}
}
