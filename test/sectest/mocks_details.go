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
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"time"
)

type SecurityMockOptions func(d *SecurityDetailsMock)

type SecurityDetailsMock struct {
	Username                 string
	UserId                   string
	AccountType              security.AccountType
	TenantExternalId         string
	TenantId                 string
	ProviderName             string
	ProviderId               string
	ProviderDisplayName      string
	ProviderDescription      string
	ProviderEmail            string
	ProviderNotificationType string
	AccessToken              string
	Exp                      time.Time
	Iss                      time.Time
	Permissions              utils.StringSet
	Roles                    utils.StringSet
	Tenants                  utils.StringSet
	OrigUsername             string
	UserFirstName            string
	UserLastName             string
	KVs                      map[string]interface{}
	ClientID                 string
	Scopes                   utils.StringSet
	OAuth2GrantType          string
	OAuth2ResponseTypes      utils.StringSet
	OAuth2Parameters         map[string]string
	OAuth2Extensions         map[string]interface{}
}

// MockedSecurityDetails implements
// - security.AuthenticationDetails
// - security.ProxiedUserDetails
// - security.UserDetails
// - security.TenantDetails
// - security.ProviderDetails
// - security.KeyValueDetails
// - internal.TenantAccessDetails
type MockedSecurityDetails struct {
	SecurityDetailsMock
}

func NewMockedSecurityDetails(opts ...SecurityMockOptions) *MockedSecurityDetails {
	ret := MockedSecurityDetails{
		SecurityDetailsMock{
			AccountType: security.AccountTypeDefault,
			ClientID: "mock",
		},
	}
	for _, fn := range opts {
		fn(&ret.SecurityDetailsMock)
	}
	return &ret
}

func (d *MockedSecurityDetails) Value(s string) (interface{}, bool) {
	v, ok := d.KVs[s]
	return v, ok
}

func (d *MockedSecurityDetails) Values() map[string]interface{} {
	return d.KVs
}

func (d *MockedSecurityDetails) OriginalUsername() string {
	return d.OrigUsername
}

func (d *MockedSecurityDetails) Proxied() bool {
	return d.OrigUsername != ""
}

func (d *MockedSecurityDetails) ExpiryTime() time.Time {
	return d.Exp
}

func (d *MockedSecurityDetails) IssueTime() time.Time {
	return d.Iss
}

func (d *MockedSecurityDetails) Roles() utils.StringSet {
	return d.SecurityDetailsMock.Roles
}

func (d *MockedSecurityDetails) Permissions() utils.StringSet {
	if d.SecurityDetailsMock.Permissions == nil {
		d.SecurityDetailsMock.Permissions = utils.NewStringSet()
	}
	return d.SecurityDetailsMock.Permissions
}

func (d *MockedSecurityDetails) AuthenticationTime() time.Time {
	return d.Iss
}

func (d *MockedSecurityDetails) ProviderId() string {
	return d.SecurityDetailsMock.ProviderId
}

func (d *MockedSecurityDetails) ProviderName() string {
	return d.SecurityDetailsMock.ProviderName
}

func (d *MockedSecurityDetails) ProviderDisplayName() string {
	return d.SecurityDetailsMock.ProviderDisplayName
}

func (d *MockedSecurityDetails) ProviderDescription() string {
	return d.SecurityDetailsMock.ProviderDescription
}

func (d *MockedSecurityDetails) ProviderEmail() string {
	return d.SecurityDetailsMock.ProviderEmail
}

func (d *MockedSecurityDetails) ProviderNotificationType() string {
	return d.SecurityDetailsMock.ProviderNotificationType
}

func (d *MockedSecurityDetails) TenantId() string {
	return d.SecurityDetailsMock.TenantId
}

func (d *MockedSecurityDetails) TenantExternalId() string {
	return d.SecurityDetailsMock.TenantExternalId
}

func (d *MockedSecurityDetails) TenantSuspended() bool {
	return false
}

func (d *MockedSecurityDetails) UserId() string {
	return d.SecurityDetailsMock.UserId
}

func (d *MockedSecurityDetails) Username() string {
	return d.SecurityDetailsMock.Username
}

func (d *MockedSecurityDetails) AccountType() security.AccountType {
	return d.SecurityDetailsMock.AccountType
}

// Deprecated: the interface is deprecated
func (d *MockedSecurityDetails) AssignedTenantIds() utils.StringSet {
	return d.EffectiveAssignedTenantIds()
}

func (d *MockedSecurityDetails) EffectiveAssignedTenantIds() utils.StringSet {
	if d.Tenants == nil {
		d.Tenants = utils.NewStringSet()
	}
	return d.Tenants
}

func (d *MockedSecurityDetails) LocaleCode() string {
	return valueFromMap[string](d.KVs, "LocaleCode")
}

func (d *MockedSecurityDetails) CurrencyCode() string {
	return valueFromMap[string](d.KVs, "CurrencyCode")
}

func (d *MockedSecurityDetails) FirstName() string {
	return d.UserFirstName
}

func (d *MockedSecurityDetails) LastName() string {
	return d.UserLastName
}

func (d *MockedSecurityDetails) Email() string {
	return valueFromMap[string](d.KVs, "Email")
}

func valueFromMap[T any](m map[string]interface{}, key string) T {
	var zero T
	if m == nil {
		return zero
	}
	if v, ok := m[key].(T); ok {
		return v
	}
	return zero
}
