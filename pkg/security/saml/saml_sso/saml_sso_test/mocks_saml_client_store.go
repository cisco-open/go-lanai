package saml_sso_test

import (
	"context"
	saml_auth_ctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/saml_sso_ctx"
	"errors"
)

type MockSamlClientStore struct {
	details []saml_auth_ctx.SamlClient
}

func NewMockedSamlClientStore(mockedClient...saml_auth_ctx.SamlClient) saml_auth_ctx.SamlClientStore {
	return &MockSamlClientStore{
		details: mockedClient}
}

func (t *MockSamlClientStore) GetAllSamlClient(_ context.Context) ([]saml_auth_ctx.SamlClient, error) {
	var result []saml_auth_ctx.SamlClient
	for _, v := range t.details {
		result = append(result, v)
	}
	return result, nil
}

func (t *MockSamlClientStore) GetSamlClientByEntityId(_ context.Context, id string) (saml_auth_ctx.SamlClient, error) {
	for _, detail := range t.details {
		if detail.GetEntityId() == id {
			return detail, nil
		}
	}
	return nil, errors.New("not found")
}