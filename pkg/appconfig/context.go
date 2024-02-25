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

package appconfig

import "github.com/cisco-open/go-lanai/pkg/bootstrap"

const (
	PropertyKeyActiveProfiles       = "application.profiles.active"
	PropertyKeyAdditionalProfiles   = "application.profiles.additional"
	PropertyKeyConfigFileSearchPath = "config.file.search-path"
	PropertyKeyApplicationName      = bootstrap.PropertyKeyApplicationName
	PropertyKeyBuildInfo            = "application.build"
	//PropertyKey = ""
)

type ConfigAccessor interface {
	bootstrap.ApplicationConfig
	Each(apply func(string, interface{}) error) error
	// Providers gives effective config providers
	Providers() []Provider
	Profiles() []string
	HasProfile(profile string) bool
}

type BootstrapConfig struct {
	config
}

func NewBootstrapConfig(groups ...ProviderGroup) *BootstrapConfig {
	return &BootstrapConfig{config: config{groups: groups}}
}

type ApplicationConfig struct {
	config
}

func NewApplicationConfig(groups ...ProviderGroup) *ApplicationConfig {
	return &ApplicationConfig{config: config{groups: groups}}
}



