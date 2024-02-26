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

package discovery

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/pkg/errors"
)

//goland:noinspection GoNameStartsWithPackageName
const (
	DiscoveryPropertiesPrefix = "cloud.consul.discovery"
)

//goland:noinspection GoNameStartsWithPackageName
type DiscoveryProperties struct {
	HealthCheckPath            string                    `json:"health-check-path"`
	HealthCheckInterval        string                    `json:"health-check-interval"`
	Tags                       utils.CommaSeparatedSlice `json:"tags"`
	AclToken                   string                    `json:"acl-token"`
	IpAddress                  string                    `json:"ip-address"` //A pre-defined IP address
	Interface                  string                    `json:"interface"`  //The network interface from where to get the ip address. If IpAddress is defined, this field is ignored
	Port                       int                       `json:"port"`
	Scheme                     string                    `json:"scheme"`
	HealthCheckCriticalTimeout string                    `json:"health-check-critical-timeout"` //See api.AgentServiceCheck's DeregisterCriticalServiceAfter field
	DefaultSelector            SelectorProperties        `json:"default-selector"`              // Default tags or meta to use when discovering other services
}

type SelectorProperties struct {
	Tags utils.CommaSeparatedSlice `json:"tags"`
	Meta map[string]string         `json:"meta"`
}

func NewDiscoveryProperties() *DiscoveryProperties {
	return &DiscoveryProperties{
		Port:                       0,
		Scheme:                     "http",
		HealthCheckInterval:        "15s",
		HealthCheckCriticalTimeout: "15s",
		HealthCheckPath:            fmt.Sprintf("%s", "/admin/health"),
	}
}

func BindDiscoveryProperties(ctx *bootstrap.ApplicationContext) DiscoveryProperties {
	props := NewDiscoveryProperties()
	if err := ctx.Config().Bind(props, DiscoveryPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DiscoveryProperties"))
	}
	return *props
}
