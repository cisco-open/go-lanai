package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"strings"
)

const (
	idPrefix       = "id-"
)

/*************************
	Account & Tenant
 *************************/

type mockedAccount struct {
	UserId          string
	Username        string
	Password        string
	DefaultTenant   string
	AssignedTenants utils.StringSet
	Permissions     utils.StringSet
}

func newMockedAccount(props *mockedAccountProperties) *mockedAccount {
	ret := &mockedAccount{
		UserId:          props.UserId,
		Username:        props.Username,
		Password:        props.Password,
		DefaultTenant:   props.DefaultTenant,
		AssignedTenants: utils.NewStringSet(props.Tenants...),
		Permissions:     utils.NewStringSet(props.Perms...),
	}
	switch {
	case ret.UserId == "":
		ret.UserId = nameToId(ret.Username)
	case ret.Username == "":
		ret.Username = idToName(ret.UserId)
	}
	return ret
}

type mockedTenant struct {
	Name string
	ID   string
}

func newMockedTenant(props *mockedTenantProperties) *mockedTenant {
	ret := &mockedTenant{
		Name: props.Name,
		ID:   props.ID,
	}
	switch {
	case ret.ID == "":
		ret.ID = nameToId(ret.Name)
	case ret.Name == "":
		ret.Name = idToName(ret.ID)
	}
	return ret
}

type mockedAccounts struct {
	idLookup map[string]*mockedAccount
	lookup   map[string]*mockedAccount
}

func newMockedAccounts(props *mockingProperties) *mockedAccounts {
	accts := mockedAccounts{
		idLookup: map[string]*mockedAccount{},
		lookup:   map[string]*mockedAccount{},
	}
	for _, v := range props.Accounts {
		acct := newMockedAccount(v)
		if acct.Username != "" {
			accts.lookup[acct.Username] = acct
		}
		if acct.UserId != "" {
			accts.idLookup[acct.UserId] = acct
		}
	}
	return &accts
}

func (m mockedAccounts) find(username, userId string) *mockedAccount {
	if v, ok := m.lookup[username]; ok && (userId == "" || v.UserId == userId) {
		return v
	}

	if v, ok := m.idLookup[userId]; ok && (username == "" || v.Username == username) {
		return v
	}
	return nil
}

func (m mockedAccounts) idToName(id string) string {
	if u, ok := m.idLookup[id]; ok {
		return u.Username
	}
	return idToName(id)
}

func (m mockedAccounts) nameToId(name string) string {
	if u, ok := m.lookup[name]; ok {
		return u.UserId
	}
	return nameToId(name)
}

type mockedTenants struct {
	idLookup map[string]*mockedTenant
	nameLookup map[string]*mockedTenant
}

func newMockedTenants(props *mockingProperties) *mockedTenants {
	tenants := mockedTenants{
		idLookup: map[string]*mockedTenant{},
		nameLookup:   map[string]*mockedTenant{},
	}
	for _, v := range props.Tenants {
		t := newMockedTenant(v)
		if t.Name != "" {
			tenants.nameLookup[t.Name] = t
		}
		if t.ID != "" {
			tenants.idLookup[t.ID] = t
		}
	}
	return &tenants
}

func (m mockedTenants) find(tenantId, tenantName string) *mockedTenant {
	if v, ok := m.idLookup[tenantId]; ok && (tenantName == "" || v.Name == tenantName) {
		return v
	}

	if v, ok := m.nameLookup[tenantName]; ok && (tenantId == "" || v.ID == tenantId) {
		return v
	}
	return nil
}

func (m mockedTenants) idToName(id string) string {
	if t, ok := m.idLookup[id]; ok {
		return t.Name
	}
	return idToName(id)
}

func (m mockedTenants) nameToId(name string) string {
	if t, ok := m.nameLookup[name]; ok {
		return t.ID
	}
	return nameToId(name)
}

/*************************
	Helpers
 *************************/

func idToName(id string) string {
	return strings.TrimPrefix(id, idPrefix)
}

func nameToId(name string) string {
	return idPrefix + name
}