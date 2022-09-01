package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

type TestFedAccountStore struct {
}

func NewTestFedAccountStore() *TestFedAccountStore {
	return &TestFedAccountStore{}
}

// LoadAccountByExternalId The externalIdName and value matches the test assertion
// The externalIdp matches that from the TestIdpManager
func (t *TestFedAccountStore) LoadAccountByExternalId(_ context.Context, externalIdName string, externalIdValue string, externalIdpName string, _ security.AutoCreateUserDetails, _ interface{}) (security.Account, error) {
	if externalIdName == "email" && externalIdValue == "test@example.com" && externalIdpName == "okta" {
		return security.NewUsernamePasswordAccount(&security.AcctDetails{
			ID:              "test@example.com",
			Type:            security.AccountTypeFederated,
			Username:        "test"}), nil
	}
	return nil, nil
}
