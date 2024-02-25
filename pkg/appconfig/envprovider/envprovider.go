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

package envprovider

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/appconfig"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"os"
	"strings"
)

type ConfigProvider struct {
	appconfig.ProviderMeta
}

const dot = rune('.')

func (configProvider *ConfigProvider) Name() string {
	return "environment-variable"
}

func (configProvider *ConfigProvider) Load(_ context.Context) (loadError error) {
	defer func() {
		if loadError != nil {
			configProvider.Loaded = false
		} else {
			configProvider.Loaded = true
		}
	}()

	flatSettings := make(map[string]interface{})

	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		k := kv[0]
		v := kv[1]

		var runes []rune
		for pos, char := range k {
			if strings.Compare(string(char), "_") == 0 {
				if pos>0 && strings.Compare(string(runes[pos-1]) , "_") != 0 {
					runes = append(runes, dot)
				} else {
					runes = append(runes, char)
				}
			} else {
				runes = append(runes, char)
			}
		}

		flatSettings[string(runes)] = utils.ParseString(v)
	}

	unFlattenedSettings, loadError := appconfig.UnFlatten(flatSettings)
	if loadError != nil {
		return loadError
	}

	configProvider.Settings = unFlattenedSettings
	return nil
}

func NewEnvProvider(precedence int) *ConfigProvider {
	return &ConfigProvider{
		ProviderMeta: appconfig.ProviderMeta{Precedence: precedence},
	}
}
