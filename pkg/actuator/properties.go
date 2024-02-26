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

package actuator

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/pkg/errors"
)

const (
	ManagementPropertiesPrefix = "management"
)

type ManagementProperties struct {
	Enabled       bool                               `json:"enabled"`
	Endpoints     EndpointsProperties                `json:"endpoints"`
	BasicEndpoint map[string]BasicEndpointProperties `json:"endpoint"`
	Security      SecurityProperties                 `json:"security"`
}

type EndpointsProperties struct {
	EnabledByDefault bool                   `json:"enabled-by-default"`
	Web              WebEndpointsProperties `json:"web"`
}

type WebEndpointsProperties struct {
	BasePath string                `json:"base-path"`
	Mappings map[string]string     `json:"path-mapping"`
	Exposure WebExposureProperties `json:"exposure"`
}

type WebExposureProperties struct {
	// Endpoint IDs that should be included or '*' for all.
	Include utils.StringSet `json:"include"`
	// Endpoint IDs that should be excluded or '*' for all.
	Exclude utils.StringSet `json:"exclude"`
}

type BasicEndpointProperties struct {
	Enabled *bool `json:"enabled"`
}

type SecurityProperties struct {
	EnabledByDefault bool                                  `json:"enabled-by-default"`
	Permissions      utils.CommaSeparatedSlice             `json:"permissions"`
	Endpoints        map[string]EndpointSecurityProperties `json:"endpoint"`
}

type EndpointSecurityProperties struct {
	Enabled     *bool                     `json:"enabled"`
	Permissions utils.CommaSeparatedSlice `json:"permissions"`
}

//NewManagementProperties create a ManagementProperties with default values
func NewManagementProperties() *ManagementProperties {
	return &ManagementProperties{
		Enabled: true,
		Endpoints: EndpointsProperties{
			Web: WebEndpointsProperties{
				BasePath: "/manage",
				Mappings: map[string]string{},
				Exposure: WebExposureProperties{
					Include: utils.NewStringSet("*"),
					Exclude: utils.NewStringSet(),
				},
			},
		},
		Security: SecurityProperties{
			EnabledByDefault: false,
			Permissions:      []string{},
			Endpoints:        map[string]EndpointSecurityProperties{
				"alive": {
					Enabled: utils.ToPtr(false),
				},
				"info": {
					Enabled: utils.ToPtr(false),
				},
				"health": {
					Enabled: utils.ToPtr(false),
				},
			},
		},
		BasicEndpoint: map[string]BasicEndpointProperties{},
	}
}

//BindManagementProperties create and bind SessionProperties, with a optional prefix
func BindManagementProperties(ctx *bootstrap.ApplicationContext) ManagementProperties {
	props := NewManagementProperties()
	if err := ctx.Config().Bind(props, ManagementPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind ManagementProperties"))
	}
	return *props
}
