package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"errors"
)

type MockAccountStore struct {
	accountLookupByUsername map[string]*mockedAccount
	accountLookupById       map[interface{}]*mockedAccount
}

func NewMockedAccountStore(props... *MockedAccountProperties) *MockAccountStore {
	store := &MockAccountStore{
		accountLookupById: make(map[interface{}]*mockedAccount),
		accountLookupByUsername: make(map[string]*mockedAccount),
	}
	for _, v := range props {
		acct := newMockedAccount(v)
		if acct.Username() != "" {
			store.accountLookupByUsername[acct.Username()] = acct
		}
		if acct.UserId != "" {
			store.accountLookupById[acct.UserId] = acct
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

func (m *MockAccountStore) LoadLockingRules(ctx context.Context, acct security.Account) (security.AccountLockingRule, error) {
	panic("implement me")
}

func (m *MockAccountStore) LoadPwdAgingRules(ctx context.Context, acct security.Account) (security.AccountPwdAgingRule, error) {
	panic("implement me")
}

func (m *MockAccountStore) Save(ctx context.Context, acct security.Account) error {
	panic("implement me")
}
