package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

type MockedTenantStore struct {
	idLookup    map[string]*security.Tenant
	extIdLookup map[string]*security.Tenant
}

func NewMockedTenantStore(props ...*MockedTenantProperties) *MockedTenantStore {
	ret := MockedTenantStore{
		idLookup:    map[string]*security.Tenant{},
		extIdLookup: map[string]*security.Tenant{},
	}
	for _, v := range props {
		t := newTenant(v)
		if len(t.ExternalId) != 0 {
			ret.extIdLookup[t.ExternalId] = t
		}
		if len(t.Id) != 0 {
			ret.idLookup[t.Id] = t
		}
	}
	return &ret
}

func newTenant(props *MockedTenantProperties) *security.Tenant {
	return &security.Tenant{
		Id:           props.ID,
		ExternalId:   props.ExternalId,
		DisplayName:  props.ExternalId,
		Description:  fmt.Sprintf("mocked tenant %s-%s", props.ID, props.ExternalId),
		ProviderId:   MockedProviderID,
		LocaleCode:   "en_US",
	}
}

func (s *MockedTenantStore) LoadTenantById(_ context.Context, id string) (*security.Tenant, error) {
	if t, ok := s.idLookup[id]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("cannot find tenant with ID [%s]", id)
}

func (s *MockedTenantStore) LoadTenantByExternalId(_ context.Context, name string) (*security.Tenant, error) {
	if t, ok := s.extIdLookup[name]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("cannot find tenant with external ID [%s]", name)
}
