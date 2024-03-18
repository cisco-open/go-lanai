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

package consulsd

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/utils/matcher"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "cloud.discovery.consul"
)

//goland:noinspection GoNameStartsWithPackageName
type DiscoveryProperties struct {
	HealthCheckScheme          string                    `json:"health-check-scheme"`
	HealthCheckPath            string                    `json:"health-check-path"`
	HealthCheckPort            int                       `json:"health-check-port"`
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
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DiscoveryProperties"))
	}
	return *props
}

// InstanceWithProperties returns an InstanceMatcher that matches instances described in given selector properties
// could return nil
func InstanceWithProperties(props *SelectorProperties) discovery.InstanceMatcher {
	matchers := make([]matcher.Matcher, 0, len(props.Tags)+len(props.Meta))
	for _, tag := range props.Tags {
		if len(tag) != 0 {
			matchers = append(matchers, discovery.InstanceWithTag(tag, true))
		}
	}
	for k, v := range props.Meta {
		matchers = append(matchers, discovery.InstanceWithMetaKV(k, v))
	}

	if len(matchers) == 0 {
		return nil
	}
	return matcher.And(matchers[0], matchers[1:]...)
}
