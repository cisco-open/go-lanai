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

package service

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security/idp"
    "github.com/cisco-open/go-lanai/pkg/security/idp/passwdidp"
)

var (
	globalLocalhostIdp = passwdidp.NewIdentityProvider(func(opt *passwdidp.PasswdIdpDetails) {
		opt.Domain = "localhost"
	})
)

// InMemoryIdpManager implements idp.IdentityProviderManager, samllogin.SamlIdentityProviderManager
type InMemoryIdpManager struct{}

// idp.IdentityProviderManager
func (i *InMemoryIdpManager) GetIdentityProvidersWithFlow(ctx context.Context, flow idp.AuthenticationFlow) []idp.IdentityProvider {
	switch flow {
	case idp.ExternalIdpSAML:
		return []idp.IdentityProvider{}
	case idp.InternalIdpForm:
		return []idp.IdentityProvider{
			globalLocalhostIdp,
		}
	}
	return []idp.IdentityProvider{}
}

// idp.IdentityProviderManager
func (i *InMemoryIdpManager) GetIdentityProviderByDomain(ctx context.Context, domain string) (idp.IdentityProvider, error) {
	switch {
	case domain == globalLocalhostIdp.Domain():
		return globalLocalhostIdp, nil
	}
	return nil, fmt.Errorf("cannot find IDP with domain %s", domain)
}

// samllogin.SamlIdentityProviderManager
func (i *InMemoryIdpManager) GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error) {
	return nil, fmt.Errorf("cannot find IDP with entity ID %s", entityId)
}

func NewInMemoryIdpManager() idp.IdentityProviderManager {
	return &InMemoryIdpManager{}
}
