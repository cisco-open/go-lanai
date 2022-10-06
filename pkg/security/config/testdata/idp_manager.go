package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/samlidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samltest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
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

func NewMockedIDPManager() *samltest.MockedIdpManager {
	return samltest.NewMockedIdpManager(func(opt *samltest.IdpManagerMockOption) {
		opt.IDPList = []idp.IdentityProvider {
			samlidp.NewIdentityProvider(func(opt *samlidp.SamlIdpDetails) {
				opt.EntityId = ExtSamlIdpEntityID
				opt.Domain = IdpDomainExtSAML
				opt.ExternalIdpName = ExtSamlIdpName
				opt.ExternalIdName = "username"
				opt.MetadataLocation = "testdata/ext-saml-metadata.xml"
			}),
		}
		opt.Delegates = []idp.IdentityProviderManager{
			sectest.NewMockedIDPManager(func(opt *sectest.IdpManagerMockOption) {
				opt.PasswdIDPDomain = IdpDomainPasswd
			}),
		}
	})
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
