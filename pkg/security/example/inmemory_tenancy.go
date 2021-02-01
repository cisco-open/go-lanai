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
	names map[string]*security.Tenant
}

func NewTenantStore() security.TenantStore {
	return &MockedTenantStore{
		ids: map[string]*security.Tenant{},
		names: map[string]*security.Tenant{},
	}
}


func (s *MockedTenantStore) LoadTenantById(ctx context.Context, id string) (*security.Tenant, error) {
	if tenant,ok := s.ids[id]; ok {
		return tenant, nil
	}
	name := fmt.Sprintf("name-for-%s", id)
	return s.new(id, name), nil
}

func (s *MockedTenantStore) LoadTenantByName(ctx context.Context, name string) (*security.Tenant, error) {
	if tenant,ok := s.names[name]; ok {
		return tenant, nil
	}
	id := fmt.Sprintf("id-for-%s", name)
	return s.new(id, name), nil
}

func (s *MockedTenantStore) new(id, name string) *security.Tenant {
	tenant := security.Tenant{
		Id:          id,
		Name:        name,
		DisplayName: name,
		Description: fmt.Sprintf("This is a mocked tenant id=%s, name=%s", id, name),
		ProviderId:  fmt.Sprintf("provider-%s", id),
		Suspended:   false,
	}
	s.ids[id] = &tenant
	s.names[name] = &tenant
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

