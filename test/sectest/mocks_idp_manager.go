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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security/idp"
)

type MockedPasswdIdentityProvider struct {
	domain string
}

func NewMockedPasswdIdentityProvider(domain string) *MockedPasswdIdentityProvider {
	return &MockedPasswdIdentityProvider{
		domain: domain,
	}
}

func (s MockedPasswdIdentityProvider) AuthenticationFlow() idp.AuthenticationFlow {
	return idp.InternalIdpForm
}

func (s MockedPasswdIdentityProvider) Domain() string {
	return s.domain
}

type MockedIDPManager struct {
	idpPasswd idp.IdentityProvider
}

type IdpManagerMockOptions func(opt *IdpManagerMockOption)
type IdpManagerMockOption struct {
	PasswdIDPDomain string
}

func NewMockedIDPManager(opts...IdpManagerMockOptions) *MockedIDPManager {
	opt := IdpManagerMockOption{}
	for _, fn := range opts {
		fn(&opt)
	}
	return &MockedIDPManager{
		idpPasswd: NewMockedPasswdIdentityProvider(opt.PasswdIDPDomain),
	}
}

func (m *MockedIDPManager) GetIdentityProvidersWithFlow(ctx context.Context, flow idp.AuthenticationFlow) []idp.IdentityProvider {
	switch flow {
	case idp.InternalIdpForm:
		return []idp.IdentityProvider{m.idpPasswd}
	default:
		return []idp.IdentityProvider{}
	}
}

func (m *MockedIDPManager) GetIdentityProviderByDomain(ctx context.Context, domain string) (idp.IdentityProvider, error) {
	switch domain {
	case m.idpPasswd.Domain():
		return m.idpPasswd, nil
	}
	return nil, fmt.Errorf("cannot find IDP for domain [%s]", domain)
}
