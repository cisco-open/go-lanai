package service

import (
    "context"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    th_loader "github.com/cisco-open/go-lanai/pkg/tenancy/loader"
    "go.uber.org/fx"
)

const (
	staticRootTenantId = "root-tenant-id"
	staticProviderId   = "static-provider-id"
)

type TenantRelation struct {
	id       string
	parentId string
}

func (t TenantRelation) GetId() string {
	return t.id
}

func (t TenantRelation) GetParentId() string {
	return t.parentId
}

type MockedTenantStore struct {
	ids         map[string]*security.Tenant
	externalIds map[string]*security.Tenant
	tenants     []th_loader.Tenant
	index       int
}

type tsDIOut struct {
	fx.Out
	HierarchyStore th_loader.TenantHierarchyStore
	TenantStore    security.TenantStore
}

func NewTenantStore(properties TenantProperties) tsDIOut {
	store := &MockedTenantStore{
		ids:         map[string]*security.Tenant{},
		externalIds: map[string]*security.Tenant{},
		tenants:     []th_loader.Tenant{TenantRelation{id: staticRootTenantId}},
	}

	for _, p := range properties.Tenants {
		t := &security.Tenant{
			Id:          p.ID,
			ExternalId:  p.ExternalId,
			DisplayName: p.Name,
			ProviderId:  staticProviderId,
		}

		store.ids[t.Id] = t
		store.externalIds[t.ExternalId] = t
		store.tenants = append(store.tenants, TenantRelation{
			id:       t.Id,
			parentId: staticRootTenantId,
		})
	}

	return tsDIOut{
		HierarchyStore: store,
		TenantStore:    store,
	}
}

func (s *MockedTenantStore) LoadTenantById(ctx context.Context, id string) (*security.Tenant, error) {
	if tenant, ok := s.ids[id]; ok {
		return tenant, nil
	}
	name := fmt.Sprintf("name-for-%s", id)
	return s.new(id, name, name), nil
}

func (s *MockedTenantStore) LoadTenantByExternalId(ctx context.Context, externalId string) (*security.Tenant, error) {
	if tenant, ok := s.externalIds[externalId]; ok {
		return tenant, nil
	}
	id := fmt.Sprintf("id-for-%s", externalId)
	return s.new(id, externalId, externalId), nil
}

func (s *MockedTenantStore) GetIterator(ctx context.Context) (th_loader.TenantIterator, error) {
	return s, nil
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

func (s *MockedTenantStore) Next() bool {
	return s.index < len(s.tenants)
}

func (s *MockedTenantStore) Scan(ctx context.Context) (th_loader.Tenant, error) {
	t := s.tenants[s.index]
	s.index++
	return t, nil
}

func (s *MockedTenantStore) Close() error {
	return nil
}

func (s *MockedTenantStore) Err() error {
	return nil
}

type StaticProviderStore struct {
}

func NewProviderStore() security.ProviderStore {
	return &StaticProviderStore{}
}

func (s *StaticProviderStore) LoadProviderById(ctx context.Context, id string) (*security.Provider, error) {
	if id != staticProviderId {
		return nil, errors.New("unknown provider Id")
	}
	return s.new(id), nil
}

func (s *StaticProviderStore) new(id string) *security.Provider {
	provider := security.Provider{
		Id:          id,
		Name:        fmt.Sprintf("name-for-%s", id),
		DisplayName: fmt.Sprintf("name-for-%s", id),
		Description: fmt.Sprintf("static provider id=%s", id),
	}
	return &provider
}
