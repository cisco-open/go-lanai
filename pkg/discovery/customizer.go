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
    "context"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/hashicorp/consul/api"
    "strings"
)

type Customizer interface {
	Customize(ctx context.Context, reg *api.AgentServiceRegistration)
}

type CustomizerFunc func(ctx context.Context, reg *api.AgentServiceRegistration)

func (fn CustomizerFunc) Customize(ctx context.Context, reg *api.AgentServiceRegistration) {
	fn(ctx, reg)
}

type Customizers struct {
	Customizers []Customizer
	applied bool
}

func NewCustomizers(ctx *bootstrap.ApplicationContext) *Customizers {
	return &Customizers{
		Customizers: []Customizer{NewDefaultCustomizer(ctx), buildInfoDiscoveryCustomizer{}},
	}
}

func (r *Customizers) Add(c Customizer) {
	if r.applied {
		panic(errors.New("cannot add consul registration customizer because other customization has already been applied"))
	}
	r.Customizers = append(r.Customizers, c)
}

func (r *Customizers) Apply(ctx context.Context, registration *api.AgentServiceRegistration) {
	if r.applied {
		return
	}
	defer func() {
		r.applied = true
	}()

	for _, c := range r.Customizers {
		c.(Customizer).Customize(ctx, registration)
	}
}

type buildInfoDiscoveryCustomizer struct {}

func (b buildInfoDiscoveryCustomizer) Customize(ctx context.Context, reg *api.AgentServiceRegistration) {
	attrs := map[string]string {
		TAG_VERSION: bootstrap.BuildVersion,
		TAG_BUILD_DATE_TIME: bootstrap.BuildTime,
	}

	components := strings.Split(bootstrap.BuildVersion, "-")
	if len(components) == 2 {
		attrs[TAG_BUILD_NUMBER] = components[1]
	}

	if reg.Meta == nil {
		reg.Meta = map[string]string{}
	}

	for k, v := range attrs {
		reg.Meta[k] = v
		reg.Tags = append(reg.Tags, fmt.Sprintf("%s=%s", k, v))
	}
}