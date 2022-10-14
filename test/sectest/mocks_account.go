package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"strings"
)

const (
	idPrefix = "id-"
)

/*************************
	Account Auth
 *************************/

type MockedAccountAuthentication struct {
	Account    MockedAccount
	AuthState  security.AuthenticationState
	DetailsMap map[string]interface{}
}

func (a MockedAccountAuthentication) Principal() interface{} {
	return a.Account
}

func (a MockedAccountAuthentication) Permissions() security.Permissions {
	perms := security.Permissions{}
	for perm := range a.Account.MockedAccountDetails.Permissions {
		perms[perm] = struct{}{}
	}
	return perms
}

func (a MockedAccountAuthentication) State() security.AuthenticationState {
	return a.AuthState
}

func (a MockedAccountAuthentication) Details() interface{} {
	return a.DetailsMap
}

/*************************
	Account
 *************************/

type MockedAccountDetails struct {
	UserId          string
	Type            security.AccountType
	Username        string
	Password        string
	TenantId        string
	DefaultTenant   string
	AssignedTenants utils.StringSet
	Permissions     utils.StringSet
}

type MockedAccount struct {
	MockedAccountDetails
}

func (m MockedAccount) DefaultDesignatedTenantId() string {
	return m.DefaultTenant
}

func (m MockedAccount) DesignatedTenantIds() []string {
	return m.AssignedTenants.Values()
}

func (m MockedAccount) TenantId() string {
	return m.MockedAccountDetails.TenantId
}

func (m MockedAccount) ID() interface{} {
	return m.UserId
}

func (m MockedAccount) Type() security.AccountType {
	return m.MockedAccountDetails.Type
}

func (m MockedAccount) Username() string {
	return m.MockedAccountDetails.Username
}

func (m MockedAccount) Credentials() interface{} {
	return ""
}

func (m MockedAccount) Permissions() []string {
	return m.MockedAccountDetails.Permissions.Values()
}

func (m MockedAccount) Disabled() bool {
	return false
}

func (m MockedAccount) Locked() bool {
	return false
}

func (m MockedAccount) UseMFA() bool {
	return false
}

func (m MockedAccount) CacheableCopy() security.Account {
	return m
}

func newMockedAccount(props *MockedAccountProperties) *MockedAccount {
	ret := &MockedAccount{
		MockedAccountDetails{
			UserId:          props.UserId,
			Type:            security.AccountTypeApp,
			Username:        props.Username,
			Password:        props.Password,
			DefaultTenant:   props.DefaultTenant,
			AssignedTenants: utils.NewStringSet(props.Tenants...),
			Permissions:     utils.NewStringSet(props.Perms...),
		},
	}
	switch {
	case ret.UserId == "":
		ret.UserId = extIdToId(ret.MockedAccountDetails.Username)
	case ret.MockedAccountDetails.Username == "":
		ret.MockedAccountDetails.Username = idToExtId(ret.UserId)
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
		if acct.MockedAccountDetails.Username != "" {
			accts.lookup[acct.MockedAccountDetails.Username] = acct
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

	if v, ok := m.idLookup[userId]; ok && (username == "" || v.MockedAccountDetails.Username == username) {
		return v
	}
	return nil
}

func (m mockedAccounts) idToName(id string) string {
	if u, ok := m.idLookup[id]; ok {
		return u.MockedAccountDetails.Username
	}
	return idToExtId(id)
}

func (m mockedAccounts) nameToId(name string) string {
	if u, ok := m.lookup[name]; ok {
		return u.UserId
	}
	return extIdToId(name)
}

type mockedTenant struct {
	ExtId string
	ID    string
}

func newMockedTenant(props *MockedTenantProperties) *mockedTenant {
	ret := &mockedTenant{
		ExtId: props.ExternalId,
		ID:    props.ID,
	}
	switch {
	case ret.ID == "":
		ret.ID = extIdToId(ret.ExtId)
	case ret.ExtId == "":
		ret.ExtId = idToExtId(ret.ID)
	}
	return ret
}

/*************************
	Tenant
 *************************/

type mockedTenants struct {
	idLookup    map[string]*mockedTenant
	extIdLookup map[string]*mockedTenant
}

func newMockedTenants(props *mockingProperties) *mockedTenants {
	tenants := mockedTenants{
		idLookup:    map[string]*mockedTenant{},
		extIdLookup: map[string]*mockedTenant{},
	}
	for _, v := range props.Tenants {
		t := newMockedTenant(v)
		if t.ExtId != "" {
			tenants.extIdLookup[t.ExtId] = t
		}
		if t.ID != "" {
			tenants.idLookup[t.ID] = t
		}
	}
	return &tenants
}

func (m mockedTenants) find(tenantId, tenantExternalId string) *mockedTenant {
	if v, ok := m.idLookup[tenantId]; ok && (tenantExternalId == "" || v.ExtId == tenantExternalId) {
		return v
	}

	if v, ok := m.extIdLookup[tenantExternalId]; ok && (tenantId == "" || v.ID == tenantId) {
		return v
	}
	return nil
}

func (m mockedTenants) idToExtId(id string) string {
	if t, ok := m.idLookup[id]; ok {
		return t.ExtId
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
