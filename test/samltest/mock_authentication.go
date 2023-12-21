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

package samltest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/crewjam/saml"
)

type MockedSamlAssertionAuthentication struct {
	Account       security.Account
	DetailsMap    map[string]interface{}
	SamlAssertion *saml.Assertion
}

func (sa *MockedSamlAssertionAuthentication) Principal() interface{} {
	return sa.Account
}

func (sa *MockedSamlAssertionAuthentication) Permissions() security.Permissions {
	perms := security.Permissions{}
	for _, perm := range sa.Account.Permissions() {
		perms[perm] = struct{}{}
	}
	return perms
}

func (sa *MockedSamlAssertionAuthentication) State() security.AuthenticationState {
	return security.StateAuthenticated
}

func (sa *MockedSamlAssertionAuthentication) Details() interface{} {
	return sa.DetailsMap
}

func (sa *MockedSamlAssertionAuthentication) Assertion() *saml.Assertion {
	return sa.SamlAssertion
}
