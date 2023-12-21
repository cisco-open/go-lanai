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

package codegen

type ConfigVersion string

const (
	VersionUnknown ConfigVersion = ``
	Version1       ConfigVersion = `v1`
	Version2       ConfigVersion = `v2`
)

var DefaultVersionedConfig = VersionedConfig{
	Version:  "v1",
	ConfigV2: DefaultConfigV2,
}

type VersionedConfig struct {
	Version ConfigVersion `json:"version"`
	Config
	ConfigV2
}

type Regeneration struct {
	Default string            `json:"default"`
	Rules   map[string]string `json:"rules"`
}

type Config struct {
	Contract           string            `json:"contract"`
	ProjectName        string            `json:"projectName"`
	TemplateDirectory  string            `json:"templateDirectory"`
	RepositoryRootPath string            `json:"repositoryRootPath"`
	Regeneration       Regeneration      `json:"regeneration"`
	Regexes            map[string]string `json:"regexes"`
}

func (c Config) ToV2() *ConfigV2 {
	regenRules := make([]RegenRule, 0, len(c.Regeneration.Rules))
	for k, v := range c.Regeneration.Rules {
		regenRules = append(regenRules, RegenRule{
			Pattern: k,
			Mode:    RegenMode(v),
		})
	}
	return &ConfigV2{
		Project:    ProjectV2{
			Name:   c.ProjectName,
			Module: c.RepositoryRootPath,
		},
		Templates:  TemplatesV2{
			Path: c.TemplateDirectory,
		},
		Components: ComponentsV2{
			Contract: ContractV2{
				Path: c.Contract,
				Naming: ContractNamingV2{
					RegExps: c.Regexes,
				},
			},
			Security: DefaultConfigV2.Components.Security,
		},
		Regen:      RegenerationV2{
			Default: RegenMode(c.Regeneration.Default),
			Rules:   regenRules,
		},
	}
}