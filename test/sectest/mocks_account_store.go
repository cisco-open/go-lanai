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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"fmt"
)

type MockAccountStoreWithFinalize struct {
	MockAccountStore
}

func NewMockedAccountStoreWithFinalize(accountProps []*MockedAccountProperties, tenantProps []*MockedTenantProperties, modifiers ...MockedAccountModifier) *MockAccountStoreWithFinalize {
	return &MockAccountStoreWithFinalize{
		MockAccountStore: *NewMockedAccountStore(accountProps, tenantProps, modifiers...),
	}
}

// Finalize will read the tenant details from the security.AccountFinalizeOption and
// adjust the user permission depending on which tenant is selected.
// Note that permissions vary depending on the combination of user + tenant.
// User1 with Tenant1 can have different permissions than User2 with Tenant1.
func (m *MockAccountStoreWithFinalize) Finalize(
	ctx context.Context,
	account security.Account,
	options ...security.AccountFinalizeOptions,
) (security.Account, error) {
	var opts security.AccountFinalizeOption
	for _, option := range options {
		option(&opts)
	}

	u, ok := m.accountLookupByUsername[account.Username()]
	if !ok {
		return nil, fmt.Errorf("username: %v not found", account.Username())
	}
	ret := *u
	ret.MockedAccountDetails.DefaultTenant = account.(security.AccountTenancy).DefaultDesignatedTenantId()
	ret.MockedAccountDetails.AssignedTenants = utils.NewStringSet(account.(security.AccountTenancy).DesignatedTenantIds()...)

	if opts.Tenant == nil {
		ret.MockedAccountDetails.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
		return ret, nil
	}
	tenant, ok := m.tenantIDLookup[opts.Tenant.Id]
	if !ok {
		return nil, fmt.Errorf("tenantID: %v not found", opts.Tenant.Id)
	}
	if permissions, ok := tenant.Permissions[account.ID().(string)]; ok {
		ret.MockedAccountDetails.Permissions = utils.NewStringSet(permissions...)
	}
	return ret, nil
}

// MockedAccountModifier works with MockAccountStore. It allows tests to modify the mocked account after load
type MockedAccountModifier func(acct security.Account) security.Account

type MockAccountStore struct {
	accountLookupByUsername map[string]*MockedAccount
	accountLookupById       map[interface{}]*MockedAccount
	tenantIDLookup          map[string]*mockedTenant
	tenantExtIDLookup       map[string]*mockedTenant
	modifiers               []MockedAccountModifier
}

func NewMockedAccountStore(accountProps []*MockedAccountProperties, tenantProps []*MockedTenantProperties, modifiers ...MockedAccountModifier) *MockAccountStore {
	store := &MockAccountStore{
		accountLookupById:       make(map[interface{}]*MockedAccount),
		accountLookupByUsername: make(map[string]*MockedAccount),
		tenantIDLookup:          map[string]*mockedTenant{},
		tenantExtIDLookup:       map[string]*mockedTenant{},
		modifiers:               modifiers,
	}
	for _, v := range accountProps {
		acct := newMockedAccount(v)
		if acct.Username() != "" {
			store.accountLookupByUsername[acct.Username()] = acct
		}
		if acct.UserId != "" {
			store.accountLookupById[acct.UserId] = acct
		}
	}
	for _, v := range tenantProps {
		t := newTenant(v)
		if len(t.ExternalId) != 0 {
			store.tenantExtIDLookup[t.ExternalId] = t
		}
		if len(t.ID) != 0 {
			store.tenantIDLookup[t.ID] = t
		}
	}
	return store
}

func (m *MockAccountStore) LoadAccountById(_ context.Context, id interface{}) (security.Account, error) {
	u, ok := m.accountLookupById[id]
	if !ok {
		return nil, errors.New("user ID not found")
	}
	var acct security.Account = u
	for _, modifier := range m.modifiers {
		acct = modifier(acct)
	}
	if acct == nil {
		return nil, errors.New("user ID not found")
	}
	return acct, nil
}

func (m *MockAccountStore) LoadAccountByUsername(_ context.Context, username string) (security.Account, error) {
	u, ok := m.accountLookupByUsername[username]
	if !ok {
		return nil, errors.New("username not found")
	}
	var acct security.Account = u
	for _, modifier := range m.modifiers {
		acct = modifier(acct)
	}
	if acct == nil {
		return nil, errors.New("username not found")
	}
	return acct, nil
}

func (m *MockAccountStore) LoadLockingRules(_ context.Context, _ security.Account) (security.AccountLockingRule, error) {
	return &security.DefaultAccount{
		AcctLockingRule: security.AcctLockingRule{
			Name: "test-noop",
		},
	}, nil
}

func (m *MockAccountStore) LoadPwdAgingRules(_ context.Context, _ security.Account) (security.AccountPwdAgingRule, error) {
	return &security.DefaultAccount{
		AcctPasswordPolicy: security.AcctPasswordPolicy{
			Name: "test-noop",
		},
	}, nil
}

func (m *MockAccountStore) Save(_ context.Context, _ security.Account) error {
	return nil
}

type MockedFederatedAccountStore struct {
	mocks []*MockedFederatedUserProperties
}

func NewMockedFederatedAccountStore(props ...*MockedFederatedUserProperties) MockedFederatedAccountStore {
	if len(props) == 0 {
		props = []*MockedFederatedUserProperties{
			{
				ExtIdpName: "*",
				ExtIdName:  "*",
				ExtIdValue: "*",
			},
		}
	}
	return MockedFederatedAccountStore{mocks: props}
}

// LoadAccountByExternalId The externalIdName and value matches the test assertion
// The externalIdp matches that from the MockedIdpName
func (s MockedFederatedAccountStore) LoadAccountByExternalId(_ context.Context, extIdName string, extIdValue string, extIdpName string, _ security.AutoCreateUserDetails, _ interface{}) (security.Account, error) {
	for i := range s.mocks {
		p := s.mocks[i]
		if extIdName != p.ExtIdName && p.ExtIdName != "*" ||
			extIdValue != p.ExtIdValue && p.ExtIdValue != "*" ||
			extIdpName != p.ExtIdpName && p.ExtIdpName != "*" {
			continue
		}
		p.UserId = s.withDefault(p.UserId, fmt.Sprintf("ext-%s-%s", extIdName, extIdValue))
		acct := newMockedAccount(&p.MockedAccountProperties)
		acct.MockedAccountDetails.Type = security.AccountTypeFederated
		return acct, nil
	}
	return nil, fmt.Errorf("unable to find federated user by extIdName=%s, extIdValue=%s, extIdpName=%s", extIdName, extIdValue, extIdpName)
}

func (s MockedFederatedAccountStore) withDefault(val, defaultVal string) string {
	if len(val) == 0 {
		return defaultVal
	}
	return val
}
