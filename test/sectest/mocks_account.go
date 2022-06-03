package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"strings"
)

const (
	idPrefix       = "id-"
)

/*************************
	Account & Tenant
 *************************/

type MockedAccount struct {
	UserId          string
	username        string
	Password        string
	tenantId        string
	DefaultTenant   string
	AssignedTenants utils.StringSet
	permissions     utils.StringSet
	UserEmail 		string
}

func (m *MockedAccount) RoleNames() []string {
	panic("implement me")
}

func (m *MockedAccount) FirstName() string {
	panic("implement me")
}

func (m *MockedAccount) LastName() string {
	panic("implement me")
}

func (m *MockedAccount) Email() string {
	return m.UserEmail
}

func (m *MockedAccount) LocaleCode() string {
	panic("implement me")
}

func (m *MockedAccount) CurrencyCode() string {
	panic("implement me")
}

func (m *MockedAccount) Value(key string) interface{} {
	panic("implement me")
}

func (m *MockedAccount) DefaultDesignatedTenantId() string {
	return m.DefaultTenant
}

func (m *MockedAccount) DesignatedTenantIds() []string {
	return m.AssignedTenants.Values()
}

func (m *MockedAccount) TenantId() string {
	return m.tenantId
}

func (m *MockedAccount) ID() interface{} {
	return m.UserId
}

func (m *MockedAccount) Type() security.AccountType {
	panic("implement me")
}

func (m *MockedAccount) Username() string {
	return m.username
}

func (m *MockedAccount) Credentials() interface{} {
	panic("implement me")
}

func (m *MockedAccount) Permissions() []string {
	return m.permissions.Values()
}

func (m MockedAccount) Disabled() bool {
	panic("implement me")
}

func (m MockedAccount) Locked() bool {
	panic("implement me")
}

func (m MockedAccount) UseMFA() bool {
	panic("implement me")
}

func (m MockedAccount) CacheableCopy() security.Account {
	panic("implement me")
}

func newMockedAccount(props *MockedAccountProperties) *MockedAccount {
	ret := &MockedAccount{
		UserId:          props.UserId,
		username:        props.Username,
		Password:        props.Password,
		DefaultTenant:   props.DefaultTenant,
		AssignedTenants: utils.NewStringSet(props.Tenants...),
		permissions:     utils.NewStringSet(props.Perms...),
	}
	switch {
	case ret.UserId == "":
		ret.UserId = extIdToId(ret.username)
	case ret.username == "":
		ret.username = idToExtId(ret.UserId)
	}
	return ret
}

type mockedTenant struct {
	ExternalId string
	ID   string
}

func newMockedTenant(props *mockedTenantProperties) *mockedTenant {
	ret := &mockedTenant{
		ExternalId: props.ExternalId,
		ID:   props.ID,
	}
	switch {
	case ret.ID == "":
		ret.ID = extIdToId(ret.ExternalId)
	case ret.ExternalId == "":
		ret.ExternalId = idToExtId(ret.ID)
	}
	return ret
}

type mockedAccounts struct {
	idLookup map[string]*MockedAccount
	lookup   map[string]*MockedAccount
}

func newMockedAccounts(props *mockingProperties) *mockedAccounts {
	accts := mockedAccounts{
		idLookup: map[string]*MockedAccount{},
		lookup:   map[string]*MockedAccount{},
	}
	for _, v := range props.Accounts {
		acct := newMockedAccount(v)
		if acct.username != "" {
			accts.lookup[acct.username] = acct
		}
		if acct.UserId != "" {
			accts.idLookup[acct.UserId] = acct
		}
	}
	return &accts
}

func (m mockedAccounts) find(username, userId string) *MockedAccount {
	if v, ok := m.lookup[username]; ok && (userId == "" || v.UserId == userId) {
		return v
	}

	if v, ok := m.idLookup[userId]; ok && (username == "" || v.username == username) {
		return v
	}
	return nil
}

func (m mockedAccounts) idToName(id string) string {
	if u, ok := m.idLookup[id]; ok {
		return u.username
	}
	return idToExtId(id)
}

func (m mockedAccounts) nameToId(name string) string {
	if u, ok := m.lookup[name]; ok {
		return u.UserId
	}
	return extIdToId(name)
}

type mockedTenants struct {
	idLookup map[string]*mockedTenant
	extIdLookup map[string]*mockedTenant
}

func newMockedTenants(props *mockingProperties) *mockedTenants {
	tenants := mockedTenants{
		idLookup: map[string]*mockedTenant{},
		extIdLookup:   map[string]*mockedTenant{},
	}
	for _, v := range props.Tenants {
		t := newMockedTenant(v)
		if t.ExternalId != "" {
			tenants.extIdLookup[t.ExternalId] = t
		}
		if t.ID != "" {
			tenants.idLookup[t.ID] = t
		}
	}
	return &tenants
}

func (m mockedTenants) find(tenantId, tenantExternalId string) *mockedTenant {
	if v, ok := m.idLookup[tenantId]; ok && (tenantExternalId == "" || v.ExternalId == tenantExternalId) {
		return v
	}

	if v, ok := m.extIdLookup[tenantExternalId]; ok && (tenantId == "" || v.ID == tenantId) {
		return v
	}
	return nil
}

func (m mockedTenants) idToExtId(id string) string {
	if t, ok := m.idLookup[id]; ok {
		return t.ExternalId
	}
	return idToExtId(id)
}

func (m mockedTenants) extIdToId(name string) string {
	if t, ok := m.extIdLookup[name]; ok {
		return t.ID
	}
	return extIdToId(name)
}

/*************************
	Helpers
 *************************/

func idToExtId(id string) string {
	return strings.TrimPrefix(id, idPrefix)
}

func extIdToId(extId string) string {
	return idPrefix + extId
}