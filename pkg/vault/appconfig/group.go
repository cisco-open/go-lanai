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

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/appconfig"
	appconfiginit "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/vault"
)

type ProviderGroupOptions func(opt *ProviderGroupOption)

type ProviderGroupOption struct {
	Precedence       int
	Backend          string
	BackendVersion   int
	Path             string
	ProfileSeparator string
	VaultClient      *vault.Client
}

// NewProviderGroup create a Vault KV engine backed appconfig.ProviderGroup.
// The provider group is responsible to load application properties from Vault KV engine at paths:
// <ProviderGroupOption.Backend>/<ProviderGroupOption.Path>[<ProviderGroupOption.ProfileSeparator><any active profile>]
// e.g.
// - "secret/defaultapplication"
// - "secret/defaultapplication/prod"
// - "secret/my-service"
// - "secret/my-service/staging"
func NewProviderGroup(opts ...ProviderGroupOptions) (appconfig.ProviderGroup, error) {
	opt := ProviderGroupOption{
		Precedence:       appconfiginit.PrecedenceExternalDefaultContext,
		Backend:          DefaultBackend,
		BackendVersion:   DefaultBackendVersion,
		Path:             DefaultConfigPath,
		ProfileSeparator: DefaultProfileSeparator,
	}
	for _, fn := range opts {
		fn(&opt)
	}

	kvSecretEngine, e := NewKvSecretEngine(opt.BackendVersion, opt.Backend, opt.VaultClient)
	if e != nil {
		return nil, e
	}

	group := appconfig.NewProfileBasedProviderGroup(opt.Precedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return opt.Path
		}
		return fmt.Sprintf("%s%s%s", opt.Path, opt.ProfileSeparator, profile)
	}
	group.CreateFunc = func(name string, order int, _ bootstrap.ApplicationConfig) appconfig.Provider {
		return NewVaultKvProvider(order, name, kvSecretEngine)
	}
	return group, nil
}
