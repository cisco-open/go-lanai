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

package sectest

import (
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/pkg/errors"
    "strings"
    "time"
)

const (
	MockingPropertiesPrefix = "mocking"
)

type MockingProperties struct {
	Accounts       MockedPropertiesAccounts       `json:"accounts"`
	Tenants        MockedPropertiesTenants        `json:"tenants"`
	Clients        MockedPropertiesClients        `json:"clients"`
	FederatedUsers MockedPropertiesFederatedUsers `json:"fed-users"`
}

// BindMockingProperties is a FX provider that bind all mocked properties as MockingProperties.
// All mocked properties should be under the yaml section defined as MockingPropertiesPrefix
// e.g. "mocking.accounts" defines all account mocks
func BindMockingProperties(ctx *bootstrap.ApplicationContext) (MockingProperties, error) {
	return MockedPropertiesBinder[MockingProperties]("")(ctx)
}

type MockedProperties[T any] map[string]*T

func (p MockedProperties[T]) Values() []*T {
	values := make([]*T, 0, len(p))
	for _, v := range p {
		values = append(values, v)
	}
	return values
}

func (p MockedProperties[T]) MapValues() map[string]*T {
	return p
}

func (p *MockedProperties[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, (*map[string]*T)(p))
}

// MockedPropertiesBinder returns a FX provider that bind specific mocked properties type from the properties sub-section
// specified by "prefix". The root section prefix is defined by MockingPropertiesPrefix
// e.g. MockedPropertiesBinder[MockedPropertiesAccounts]("accounts"):
//	    The returned binder binds MockedPropertiesAccounts from "mocking.accounts"
func MockedPropertiesBinder[T any](prefix string) func(ctx *bootstrap.ApplicationContext) (T, error) {
	return func(ctx *bootstrap.ApplicationContext) (T, error) {
		prefix = MockingPropertiesPrefix + "." + prefix
		prefix = strings.Trim(prefix, ".")
		var props T
		if err := ctx.Config().Bind(&props, prefix); err != nil {
			return props, errors.Wrap(err, fmt.Sprintf("failed to bind mocking properties %T from [%s]", props, prefix))
		}
		return props, nil
	}
}

type MockedPropertiesClients struct {
	MockedProperties[MockedClientProperties]
}

type MockedClientProperties struct {
	ClientID          string                    `json:"id"`
	Secret            string                    `json:"secret"`
	GrantTypes        utils.CommaSeparatedSlice `json:"grant-types"`
	Scopes            utils.CommaSeparatedSlice `json:"scopes"`
	RedirectUris      utils.CommaSeparatedSlice `json:"redirect-uris"`
	ATValidity        utils.Duration            `json:"access-token-validity"`
	RTValidity        utils.Duration            `json:"refresh-token-validity"`
	AssignedTenantIds utils.CommaSeparatedSlice `json:"tenants"`
}

type MockedPropertiesAccounts struct {
	MockedProperties[MockedAccountProperties]
}

type MockedAccountProperties struct {
	UserId        string   `json:"id"` // optional field
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	DefaultTenant string   `json:"default-tenant"`
	Tenants       []string `json:"tenants"`
	Perms         []string `json:"permissions"`
}

type MockedPropertiesFederatedUsers struct {
	MockedProperties[MockedFederatedUserProperties]
}

type MockedFederatedUserProperties struct {
	MockedAccountProperties
	ExtIdpName string `json:"ext-idp-name"`
	ExtIdName  string `json:"ext-id-name"`
	ExtIdValue string `json:"ext-id-value"`
}

type MockedPropertiesTenants struct {
	MockedProperties[MockedTenantProperties]
}

type MockedTenantProperties struct {
	ID         string              `json:"id"` // optional field
	ExternalId string              `json:"external-id"`
	Perms      map[string][]string `json:"permissions"` // permissions are MockedAccountProperties.UserId to permissions
}

type scopeMockingProperties struct {
	MockingProperties
	TokenValidity utils.Duration `json:"token-validity"`
}

func bindScopeMockingProperties(ctx *bootstrap.ApplicationContext) *scopeMockingProperties {
	props := scopeMockingProperties{
		TokenValidity: utils.Duration(120 * time.Second),
	}
	if err := ctx.Config().Bind(&props, MockingPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind mocking properties"))
	}
	return &props
}
