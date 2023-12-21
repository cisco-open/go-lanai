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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

const (
	PropertiesPrefix = "mocking"
)

type mockingProperties struct {
	Accounts      map[string]*MockedAccountProperties `json:"accounts"`
	Tenants       map[string]*MockedTenantProperties  `json:"tenants"`
	TokenValidity utils.Duration                      `json:"token-validity"`
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

type MockedAccountProperties struct {
	UserId        string   `json:"id"` // optional field
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	DefaultTenant string   `json:"default-tenant"`
	Tenants       []string `json:"tenants"`
	Perms         []string `json:"permissions"`
}

type MockedFederatedUserProperties struct {
	MockedAccountProperties
	ExtIdpName string `json:"ext-idp-name"`
	ExtIdName  string `json:"ext-id-name"`
	ExtIdValue string `json:"ext-id-value"`
}

type MockedTenantProperties struct {
	ID         string              `json:"id"` // optional field
	ExternalId string              `json:"external-id"`
	Perms      map[string][]string `json:"permissions"` // permissions are MockedAccountProperties.UserId to permissions
}

func bindMockingProperties(ctx *bootstrap.ApplicationContext) *mockingProperties {
	props := mockingProperties{
		Accounts:      map[string]*MockedAccountProperties{},
		Tenants:       map[string]*MockedTenantProperties{},
		TokenValidity: utils.Duration(120 * time.Second),
	}
	if err := ctx.Config().Bind(&props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind mocking properties"))
	}
	return &props
}
