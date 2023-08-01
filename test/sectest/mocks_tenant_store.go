package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
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
