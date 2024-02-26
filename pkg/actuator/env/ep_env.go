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

package env

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/appconfig"
	"sort"
)

const (
	ID                   = "env"
	EnableByDefault      = false
)

type Input struct {
	Pattern string `form:"match"`
}

type EnvDescriptor struct {
	ActiveProfiles  []string            `json:"activeProfiles,omitempty"`
	PropertySources []PSourceDescriptor `json:"propertySources,omitempty"`
}

type PSourceDescriptor struct {
	Name string `json:"name"`
	Properties map[string]PValueDescriptor `json:"properties,omitempty"`
	order int
}

type PValueDescriptor struct {
	Value  interface{} `json:"value,omitempty"`
	Origin string      `json:"origin,omitempty"`
}

// EnvEndpoint implements actuator.Endpoint, actuator.WebEndpoint
type EnvEndpoint struct {
	actuator.WebEndpointBase
	appConfig appconfig.ConfigAccessor
	sanitizer *Sanitizer
}

func new(di regDI) *EnvEndpoint {
	ep := EnvEndpoint{
		appConfig: di.AppContext.Config().(appconfig.ConfigAccessor),
		sanitizer: NewSanitizer(di.Properties.KeysToSanitize.Values()),
	}
	ep.WebEndpointBase = actuator.MakeWebEndpointBase(func(opt *actuator.EndpointOption) {
		opt.Id = ID
		opt.Ops = []actuator.Operation{
			actuator.NewReadOperation(ep.Read),
		}
		opt.Properties = &di.MgtProperties.Endpoints
		opt.EnabledByDefault = EnableByDefault
	})
	return &ep
}

// Read never returns error
func (ep *EnvEndpoint) Read(ctx context.Context, input *Input) (*EnvDescriptor, error) {
	// TODO maybe support match pattern
	env := EnvDescriptor{
		ActiveProfiles: ep.appConfig.Profiles(),
		PropertySources: []PSourceDescriptor{},
	}

	for _, provider := range ep.appConfig.Providers() {
		if !provider.IsLoaded() {
			continue
		}

		psrc := PSourceDescriptor{
			Name: provider.Name(),
			Properties: map[string]PValueDescriptor{},
			order: provider.Order(),
		}

		values := provider.GetSettings()
		_ = appconfig.VisitEach(values, func(k string, v interface{}) error {
			v = ep.sanitizer.Sanitize(ctx, k, v)
			psrc.Properties[k] = PValueDescriptor{Value: v, Origin: ""}
			return nil
		})
		if len(psrc.Properties) > 0 {
			env.PropertySources = append(env.PropertySources, psrc)
		}
	}

	sort.SliceStable(env.PropertySources, func(i, j int) bool {
		return env.PropertySources[i].order < env.PropertySources[j].order
	})
	return &env, nil
}



