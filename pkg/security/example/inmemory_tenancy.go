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

package example

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

const (
	specialIdNonExist = "non-exist"
	specialNameNonExist = "non-exist"
)

// security.TenantStore
type MockedTenantStore struct {
	ids map[string]*security.Tenant
	externalIds map[string]*security.Tenant
}

func NewTenantStore() security.TenantStore {
	return &MockedTenantStore{
		ids: map[string]*security.Tenant{},
		externalIds: map[string]*security.Tenant{},
	}
}


func (s *MockedTenantStore) LoadTenantById(ctx context.Context, id string) (*security.Tenant, error) {
	if tenant,ok := s.ids[id]; ok {
		return tenant, nil
	}
	name := fmt.Sprintf("name-for-%s", id)
	return s.new(id, name, name), nil
}

func (s *MockedTenantStore) LoadTenantByExternalId(ctx context.Context, externalId string) (*security.Tenant, error) {
	if tenant,ok := s.externalIds[externalId]; ok {
		return tenant, nil
	}
	id := fmt.Sprintf("id-for-%s", externalId)
	return s.new(id, externalId, externalId), nil
}

func (s *MockedTenantStore) new(id, name string, externalId string) *security.Tenant {
	tenant := security.Tenant{
		Id:          id,
		ExternalId:  externalId,
		DisplayName: name,
		Description: fmt.Sprintf("This is a mocked tenant id=%s, name=%s", id, name),
		ProviderId:  fmt.Sprintf("provider-%s", id),
		Suspended:   false,
	}
	s.ids[id] = &tenant
	s.externalIds[externalId] = &tenant
	return &tenant
}

// security.ProviderStore
type MockedProviderStore struct {
	ids map[string]*security.Provider
}

func NewProviderStore() security.ProviderStore {
	return &MockedProviderStore{
		ids: map[string]*security.Provider{},
	}
}

func (s *MockedProviderStore) LoadProviderById(ctx context.Context, id string) (*security.Provider, error) {
	if provider,ok := s.ids[id]; ok {
		return provider, nil
	}
	return s.new(id), nil
}

func (s *MockedProviderStore) new(id string) *security.Provider {
	provider := security.Provider{
		Id: id,
		Name: fmt.Sprintf("name-for-%s", id),
		DisplayName: fmt.Sprintf("name-for-%s", id),
		Description: fmt.Sprintf("This is a mocked provider id=%s", id),
	}
	s.ids[id] = &provider
	return &provider
}

