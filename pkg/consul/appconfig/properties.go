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

package consulappconfig

import "github.com/cisco-open/go-lanai/pkg/appconfig"

const (
	PropertiesPrefix = "cloud.consul.config"
	DefaultConfigPathPrefix = "userviceconfiguration"
	DefaultConfigPath = "defaultapplication"
	DefaultProfileSeparator = ","
)

type ConsulConfigProperties struct {
	Enabled          bool   `json:"enabled"`
	Prefix           string `json:"prefix"`
	DefaultContext   string `json:"default-context"`
	ProfileSeparator string `json:"profile-separator"`
}

func bindConsulConfigProperties(bootstrapConfig *appconfig.BootstrapConfig) (ConsulConfigProperties, error) {
	p := ConsulConfigProperties{
		Prefix:           DefaultConfigPathPrefix,
		DefaultContext:   DefaultConfigPath,
		ProfileSeparator: DefaultProfileSeparator,
		Enabled:          true,
	}
	if e := bootstrapConfig.Bind(&p, PropertiesPrefix); e != nil {
		return p, e
	}
	return p, nil
}
