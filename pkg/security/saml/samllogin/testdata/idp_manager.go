package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"errors"
)

type TestIdpManager struct {
	idpDetails TestIdpProvider
}

func NewTestIdpManager() *TestIdpManager {
	return &TestIdpManager{
		idpDetails: TestIdpProvider{
			domain:           "saml.vms.com",
			metadataLocation: "testdata/okta_metadata.xml",
			externalIdpName: "okta",
			externalIdName: "email",
			entityId: "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
		},
	}
}

func (t *TestIdpManager) GetIdentityProvidersWithFlow(context.Context, idp.AuthenticationFlow) []idp.IdentityProvider {
	return []idp.IdentityProvider{t.idpDetails}
}

func (t TestIdpManager) GetIdentityProviderByEntityId(_ context.Context, entityId string) (idp.IdentityProvider, error) {
	if entityId == t.idpDetails.entityId {
		return t.idpDetails, nil
	}
	return nil, errors.New("not found")
}

func (t TestIdpManager) GetIdentityProviderByDomain(_ context.Context, domain string) (idp.IdentityProvider, error) {
	if domain == t.idpDetails.domain {
		return t.idpDetails, nil
	}
	return nil, errors.New("not found")
}
