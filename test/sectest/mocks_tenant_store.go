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
    "github.com/cisco-open/go-lanai/pkg/security"
)

type MockedTenantStore struct {
	idLookup    map[string]*mockedTenant
	extIdLookup map[string]*mockedTenant
}

func NewMockedTenantStore(props ...*MockedTenantProperties) *MockedTenantStore {
	ret := MockedTenantStore{
		idLookup:    map[string]*mockedTenant{},
		extIdLookup: map[string]*mockedTenant{},
	}
	for _, v := range props {
		t := newTenant(v)
		if len(t.ExternalId) != 0 {
			ret.extIdLookup[t.ExternalId] = t
		}
		if len(t.ID) != 0 {
			ret.idLookup[t.ID] = t
		}
	}
	return &ret
}

func newTenant(props *MockedTenantProperties) *mockedTenant {
	return &mockedTenant{
		ID:          props.ID,
		ExternalId:  props.ExternalId,
		ProviderId:  MockedProviderID,
		Permissions: props.Perms,
	}
}

func (s *MockedTenantStore) LoadTenantById(_ context.Context, id string) (*security.Tenant, error) {
	if t, ok := s.idLookup[id]; ok {
		return toSecurityTenant(t), nil
	}
	return nil, fmt.Errorf("cannot find tenant with ID [%s]", id)
}

func (s *MockedTenantStore) LoadTenantByExternalId(_ context.Context, name string) (*security.Tenant, error) {
	if t, ok := s.extIdLookup[name]; ok {
		return toSecurityTenant(t), nil
	}
	return nil, fmt.Errorf("cannot find tenant with external ID [%s]", name)
}

func toSecurityTenant(mocked *mockedTenant) *security.Tenant {
	return &security.Tenant{
		Id:          mocked.ID,
		ExternalId:  mocked.ExternalId,
		DisplayName: mocked.ExternalId,
		ProviderId:  mocked.ProviderId,
	}
}
