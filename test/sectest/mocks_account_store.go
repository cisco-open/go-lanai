package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"fmt"
)

type MockAccountStoreWithFinalize struct {
	MockAccountStore
}

func NewMockedAccountStoreWithFinalize(accountProps []*MockedAccountProperties, tenantProps []*MockedTenantProperties) *MockAccountStoreWithFinalize {
	return &MockAccountStoreWithFinalize{
		MockAccountStore: *NewMockedAccountStore(accountProps, tenantProps),
	}
}

// Finalize will read the tenant details from the security.AccountFinalizeOption and
// adjust the user permission depending on which tenant is selected.
// Note that permissions vary depending on the combination of user + tenant.
// User1 with Tenant1 can have different permissions than User2 with Tenant1.
func (m *MockAccountStoreWithFinalize) Finalize(
	ctx context.Context,
	account security.Account,
	options ...security.AccountFinalizeOptions,
) (security.Account, error) {
	var opts security.AccountFinalizeOption
	for _, option := range options {
		option(&opts)
	}

	u, ok := m.accountLookupByUsername[account.Username()]
	if !ok {
		return nil, fmt.Errorf("username: %v not found", account.Username())
	}
	if opts.Tenant == nil {
		u.MockedAccountDetails.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
		return u, nil
	}
	tenant, ok := m.tenantIDLookup[opts.Tenant.Id]
	if !ok {
		return nil, fmt.Errorf("tenantID: %v not found", opts.Tenant.Id)
	}
	if permissions, ok := tenant.Permissions[account.ID().(string)]; ok {
		u.MockedAccountDetails.Permissions = utils.NewStringSet(permissions...)
	}
	return u, nil
}

type MockAccountStore struct {
	accountLookupByUsername map[string]*MockedAccount
	accountLookupById       map[interface{}]*MockedAccount
	tenantIDLookup          map[string]*mockedTenant
	tenantExtIDLookup       map[string]*mockedTenant
}

func NewMockedAccountStore(accountProps []*MockedAccountProperties, tenantProps []*MockedTenantProperties) *MockAccountStore {
	store := &MockAccountStore{
		accountLookupById:       make(map[interface{}]*MockedAccount),
		accountLookupByUsername: make(map[string]*MockedAccount),
		tenantIDLookup:          map[string]*mockedTenant{},
		tenantExtIDLookup:       map[string]*mockedTenant{},
	}
	for _, v := range accountProps {
		acct := newMockedAccount(v)
		if acct.Username() != "" {
			store.accountLookupByUsername[acct.Username()] = acct
		}
		if acct.UserId != "" {
			store.accountLookupById[acct.UserId] = acct
		}
	}
	for _, v := range tenantProps {
		t := newTenant(v)
		if len(t.ExternalId) != 0 {
			store.tenantExtIDLookup[t.ExternalId] = t
		}
		if len(t.ID) != 0 {
			store.tenantIDLookup[t.ID] = t
		}
	}
	return store
}

func (m *MockAccountStore) LoadAccountById(_ context.Context, id interface{}) (security.Account, error) {
	u, ok := m.accountLookupById[id]
	if !ok {
		return nil, errors.New("user Domain not found")
	}

	return u, nil
}

func (m *MockAccountStore) LoadAccountByUsername(_ context.Context, username string) (security.Account, error) {
	u, ok := m.accountLookupByUsername[username]
	if !ok {
		return nil, errors.New("username not found")
	}

	return u, nil
}

func (m *MockAccountStore) LoadLockingRules(_ context.Context, _ security.Account) (security.AccountLockingRule, error) {
	return &security.DefaultAccount{
		AcctLockingRule: security.AcctLockingRule{
			Name: "test-noop",
		},
	}, nil
}

func (m *MockAccountStore) LoadPwdAgingRules(_ context.Context, _ security.Account) (security.AccountPwdAgingRule, error) {
	return &security.DefaultAccount{
		AcctPasswordPolicy: security.AcctPasswordPolicy{
			Name: "test-noop",
		},
	}, nil
}

func (m *MockAccountStore) Save(_ context.Context, _ security.Account) error {
	return nil
}

type MockedFederatedAccountStore struct {
	mocks []*MockedFederatedUserProperties
}

func NewMockedFederatedAccountStore(props ...*MockedFederatedUserProperties) MockedFederatedAccountStore {
	if len(props) == 0 {
		props = []*MockedFederatedUserProperties{
			{
				ExtIdpName: "*",
				ExtIdName:  "*",
				ExtIdValue: "*",
			},
		}
	}
	return MockedFederatedAccountStore{mocks: props}
}

// LoadAccountByExternalId The externalIdName and value matches the test assertion
// The externalIdp matches that from the MockedIdpName
func (s MockedFederatedAccountStore) LoadAccountByExternalId(_ context.Context, extIdName string, extIdValue string, extIdpName string, _ security.AutoCreateUserDetails, _ interface{}) (security.Account, error) {
	for _, p := range s.mocks {
		if extIdName != p.ExtIdName && p.ExtIdName != "*" ||
			extIdValue != p.ExtIdValue && p.ExtIdValue != "*" ||
			extIdpName != p.ExtIdpName && p.ExtIdpName != "*" {
			continue
		}
		p.UserId = s.withDefault(p.UserId, fmt.Sprintf("ext-%s-%s", extIdName, extIdValue))
		acct := newMockedAccount(&p.MockedAccountProperties)
		acct.MockedAccountDetails.Type = security.AccountTypeFederated
		return acct, nil
	}
	return nil, fmt.Errorf("unable to find federated user by extIdName=%s, extIdValue=%s, extIdpName=%s", extIdName, extIdValue, extIdpName)
}

func (s MockedFederatedAccountStore) withDefault(val, defaultVal string) string {
	if len(val) == 0 {
		return defaultVal
	}
	return val
}
