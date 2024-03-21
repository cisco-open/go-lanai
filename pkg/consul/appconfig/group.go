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

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/appconfig"
	appconfiginit "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
)

type ProviderGroupOptions func(opt *ProviderGroupOption)

type ProviderGroupOption struct {
	Precedence       int
	Prefix           string
	Path             string
	ProfileSeparator string
	Connection       *consul.Connection
}

// NewProviderGroup create a Consul KV store backed appconfig.ProviderGroup.
// The provider group is responsible to load application properties from Consul KV store at paths:
// <ProviderGroupOption.Prefix>/<ProviderGroupOption.Path>[<ProviderGroupOption.ProfileSeparator><any active profile>]
// e.g.
// - "userviceconfiguration/defaultapplication"
// - "userviceconfiguration/defaultapplication,prod"
// - "userviceconfiguration/my-service"
// - "userviceconfiguration/my-service,staging"
func NewProviderGroup(opts ...ProviderGroupOptions) appconfig.ProviderGroup {
	opt := ProviderGroupOption{
		Precedence:       appconfiginit.PrecedenceExternalDefaultContext,
		Prefix:           DefaultConfigPathPrefix,
		Path:             DefaultConfigPath,
		ProfileSeparator: DefaultProfileSeparator,
	}
	for _, fn := range opts {
		fn(&opt)
	}

	group := appconfig.NewProfileBasedProviderGroup(opt.Precedence)
	group.KeyFunc = func(profile string) string {
		if profile == "" {
			return fmt.Sprintf("%s/%s", opt.Prefix, opt.Path)
		}
		return fmt.Sprintf("%s/%s%s%s", opt.Prefix, opt.Path, opt.ProfileSeparator, profile)
	}
	group.CreateFunc = func(name string, order int, _ bootstrap.ApplicationConfig) appconfig.Provider {
		ptr := NewConfigProvider(order, name, opt.Connection)
		if ptr == nil {
			return nil
		}
		return ptr
	}
	return group
}

