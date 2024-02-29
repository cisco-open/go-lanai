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

package vaultappconfig

import "github.com/cisco-open/go-lanai/pkg/appconfig"

const (
	PropertiesPrefix = "cloud.vault.kv"
	DefaultBackend = `secret`
	DefaultBackendVersion = 1
	DefaultConfigPath = "defaultapplication"
	DefaultProfileSeparator = "/"
)

// VaultConfigProperties currently only supports v1 kv secret engine
// TODO review property path and prefix
type VaultConfigProperties struct {
	Enabled     bool `json:"enabled"`
	Backend          string `json:"backend"`
	BackendVersion   int    `json:"backend-version"`
	DefaultContext   string `json:"default-context"`
	ProfileSeparator string `json:"profile-separator"`
}

func bindVaultConfigProperties(bootstrapConfig *appconfig.BootstrapConfig) VaultConfigProperties {
	p := VaultConfigProperties{
		Enabled:          true,
		Backend:          DefaultBackend,
		BackendVersion:   DefaultBackendVersion,
		DefaultContext:   DefaultConfigPath,
		ProfileSeparator: DefaultProfileSeparator,
	}
	if e := bootstrapConfig.Bind(&p, PropertiesPrefix); e != nil {
		panic(e)
	}
	return p
}


