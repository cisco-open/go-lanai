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

package security

import "github.com/cisco-open/go-lanai/pkg/utils"

type MockedAccountAuth struct {
	permissions Permissions
	details     UserDetails
}

type MockedUserDetails struct {
	userId            string
	username          string
	assignedTenantIds utils.StringSet
}

func (m MockedUserDetails) UserId() string {
	return m.userId
}

func (m MockedUserDetails) Username() string {
	return m.username
}

func (m MockedUserDetails) AccountType() AccountType {
	return 0
}

// Deprecated: the interface is deprecated
func (m MockedUserDetails) AssignedTenantIds() utils.StringSet {
	return m.assignedTenantIds
}

func (m MockedUserDetails) EffectiveAssignedTenantIds() utils.StringSet {
	return m.assignedTenantIds
}

func (m MockedUserDetails) LocaleCode() string {
	return ""
}

func (m MockedUserDetails) CurrencyCode() string {
	return ""
}

func (m MockedUserDetails) FirstName() string {
	return ""
}

func (m MockedUserDetails) LastName() string {
	return ""
}

func (m MockedUserDetails) Email() string {
	return ""
}

func (m MockedAccountAuth) Principal() interface{} {
	return nil
}

func (m MockedAccountAuth) Permissions() Permissions {
	perms := Permissions{}
	for perm := range m.permissions {
		perms[perm] = struct{}{}
	}
	return perms
}

func (m MockedAccountAuth) State() AuthenticationState {
	return StateAuthenticated
}

func (m MockedAccountAuth) Details() interface{} {
	return m.details
}
