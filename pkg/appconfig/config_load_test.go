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

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	. "github.com/onsi/gomega"
	"testing"
)

const (
	KeyProvidersCap       = "providers-cap"
	KeyValueOverridden    = "overridden"
	KeyValueNotOverridden = "not-overridden"
	ValueNotOverridden    = "fixed value"
)

var (
	Groups = []ProviderGroup{
		&TestProviderGroup{
			order: 0,
			providers: []Provider{
				&TestProvider{
					mocked: map[string]interface{}{
						KeyProvidersCap:    2,
						KeyValueOverridden: "g-0-p-1",
						"application": map[string]interface{}{
							"profiles": map[string]interface{}{
								"additional": []string{"profile-0-1"},
								"active":     []interface{}{"active-0-1"},
							},
						},
					},
					name:  "g-0-p-1",
					order: 1,
				},
				&TestProvider{
					mocked: map[string]interface{}{
						KeyProvidersCap:    3,
						KeyValueOverridden: "g-0-p-0",
						"application": map[string]interface{}{
							"profiles": map[string]interface{}{
								"additional": []string{"profile-0-0"},
								"active":     []interface{}{"active-0-0"},
							},
						},
					},
					name:  "g-0-p-0",
					order: 0,
				},
			},
		},
		&TestProviderGroup{
			order: 1,
			providers: []Provider{
				&TestProvider{
					mocked: map[string]interface{}{
						KeyProvidersCap:       1,
						KeyValueOverridden:    "g-1-p-2",
						KeyValueNotOverridden: ValueNotOverridden,
						"application": map[string]interface{}{
							"profiles": map[string]interface{}{
								"additional": []string{"profile-1-2"},
								"active":     []interface{}{"active-1-2"},
							},
						},
					},
					name:  "g-1-p-2",
					order: 2,
				},
				&TestProvider{
					mocked: map[string]interface{}{
						KeyProvidersCap:    1,
						KeyValueOverridden: "g-1-p-0",
						"application": map[string]interface{}{
							"profiles": map[string]interface{}{
								"additional": []string{"profile-1-0"},
								"active":     []interface{}{"active-1-0"},
							},
						},
					},
					name:  "g-1-p-0",
					order: 0,
				},
				&TestProvider{
					mocked: map[string]interface{}{
						KeyProvidersCap:    1,
						KeyValueOverridden: "g-1-p-1",
						"application": map[string]interface{}{
							"profiles": map[string]interface{}{
								"additional": []string{"profile-1-1"},
								"active":     []interface{}{"active-1-1"},
							},
						},
					},
					name:  "g-1-p-1",
					order: 1,
				},
			},
		},
	}
)

func TestMultiPassLoad(t *testing.T) {
	conf := config{
		groups: Groups,
	}
	g := NewWithT(t)
	e := conf.Load(context.Background(), false)
	g.Expect(e).To(Succeed(), "Load shouldn't return error")
	g.Expect(len(conf.Providers())).To(Equal(5), "All providers should be loaded")
	first := conf.Providers()[0]
	g.Expect(conf.Value(KeyValueOverridden)).To(Equal(first.Name()), "overridden value should be set by highest precedence provider")
	g.Expect(conf.Value(KeyValueNotOverridden)).To(Equal(ValueNotOverridden), "not overridden value should be correct")
	g.Expect(conf.Value(PropertyKeyAdditionalProfiles)).To(HaveLen(5), "additional profiles should appended and have 5 entries")
	g.Expect(conf.Value(PropertyKeyActiveProfiles)).To(HaveLen(1), "active profiles should override and have 1 entry")
	ap := conf.Value(PropertyKeyActiveProfiles).([]interface{})
	g.Expect(ap[0]).To(Equal("active-0-0"))
	g.Expect(conf.profiles).To(HaveLen(6), "should have active profiles plus additional profiles")
}

/*********************
	SubTests
 *********************/

/*********************
	Helpers
 *********************/

/*
********************

		Mock Types
	 ********************
*/
type TestProviderGroup struct {
	order       int
	providers   []Provider
	shouldReset bool
}

func (g TestProviderGroup) Order() int {
	return g.order
}

func (g TestProviderGroup) Providers(_ context.Context, config bootstrap.ApplicationConfig) []Provider {
	c, ok := config.Value(KeyProvidersCap).(int)
	if !ok || c == 0 {
		c = 1
	} else if c > len(g.providers) {
		c = len(g.providers)
	}
	return g.providers[:c]
}

func (g *TestProviderGroup) Reset() {
	g.shouldReset = true
}

type TestProvider struct {
	mocked   map[string]interface{}
	settings map[string]interface{}
	name     string
	order    int
}

func (p TestProvider) Name() string {
	return p.name
}

func (p *TestProvider) Load(_ context.Context) error {
	p.settings = p.mocked
	return nil
}

func (p TestProvider) GetSettings() map[string]interface{} {
	return p.settings
}

func (p TestProvider) Order() int {
	return p.order
}

func (p TestProvider) IsLoaded() bool {
	return p.settings != nil
}

func (p *TestProvider) Reset() {
	p.settings = nil
}
