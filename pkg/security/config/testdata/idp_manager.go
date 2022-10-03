package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/samlidp"
	"fmt"
)

const (
	IdpDomainPasswd    = "passwd.lanai.com"
	IdpDomainExtSAML   = "saml.lanai.com"
	ExtSamlIdpName     = "ext-saml-idp"
	ExtSamlIdpEntityID = "http://external.saml.com/samlidp/metadata"
	ExtSamlIdpSSOUrl = "http://external.saml.com/samlidp/authorize"
	ExtSamlIdpSLOUrl = "http://external.saml.com/samlidp/logout"
)

type MockedIDPManager struct {
	idpPasswd idp.IdentityProvider
	idpSaml   idp.IdentityProvider
}

func NewMockedIDPManager() *MockedIDPManager {
	return &MockedIDPManager{
		idpPasswd: passwdidp.NewIdentityProvider(func(opt *passwdidp.PasswdIdpDetails) {
			opt.Domain = IdpDomainPasswd
		}),
		idpSaml: samlidp.NewIdentityProvider(func(opt *samlidp.SamlIdpDetails) {
			opt.EntityId = ExtSamlIdpEntityID
			opt.Domain = IdpDomainExtSAML
			opt.ExternalIdpName = ExtSamlIdpName
			opt.ExternalIdName = "username"
			opt.MetadataLocation = "testdata/ext-saml-metadata.xml"
		}),
	}
}

func (m *MockedIDPManager) GetIdentityProvidersWithFlow(_ context.Context, flow idp.AuthenticationFlow) []idp.IdentityProvider {
	switch flow {
	case idp.InternalIdpForm:
		return []idp.IdentityProvider{m.idpPasswd}
	case idp.ExternalIdpSAML:
		return []idp.IdentityProvider{m.idpSaml}
	default:
		return []idp.IdentityProvider{}
	}
}

func (m *MockedIDPManager) GetIdentityProviderByDomain(_ context.Context, domain string) (idp.IdentityProvider, error) {
	switch domain {
	case IdpDomainPasswd:
		return m.idpPasswd, nil
	case IdpDomainExtSAML:
		return m.idpSaml, nil
	default:
		return nil, fmt.Errorf("cannot find IDP for domain [%s]", domain)
	}
}

func (m *MockedIDPManager) GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error) {
	if ExtSamlIdpEntityID == entityId {
		return m.idpSaml, nil
	}
	return nil, fmt.Errorf("cannot find IDP for entityId [%s]", entityId)
}

type MockedFedAccountStore struct{}

func NewMockedFedAccountStore() MockedFedAccountStore {
	return MockedFedAccountStore{}
}

// LoadAccountByExternalId The externalIdName and value matches the test assertion
// The externalIdp matches that from the TestIdpManager
func (t MockedFedAccountStore) LoadAccountByExternalId(_ context.Context, externalIdName string, externalIdValue string, externalIdpName string, _ security.AutoCreateUserDetails, _ interface{}) (security.Account, error) {
	if externalIdpName == ExtSamlIdpName {
		return security.NewUsernamePasswordAccount(&security.AcctDetails{
			ID:       fmt.Sprintf("%s-%s", externalIdName, externalIdValue),
			Type:     security.AccountTypeFederated,
			Username: externalIdValue}), nil
	}
	return nil, nil
}
