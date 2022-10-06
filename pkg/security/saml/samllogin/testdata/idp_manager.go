package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samltest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
)

var DefaultIdpProviders = []idp.IdentityProvider {
	samltest.MockedIdpProvider {
		ExtSamlMetadata: samltest.ExtSamlMetadata{
			EntityId:         "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
			Domain:           "saml.vms.com",
			Source:           "testdata/okta_login_test_metadata.xml",
			Name:             "okta",
			IdName:           "email",
		},
	},
	samltest.MockedIdpProvider{
		ExtSamlMetadata: samltest.ExtSamlMetadata{
			EntityId:         "http://www.okta.com/exk668ha29xaI4in25d7",
			Domain:           "saml-alt.vms.com",
			Source:           "testdata/okta_logout_test_metadata.xml",
			Name:             "okta",
			IdName:           "email",
		},
	},
}

var DefaultFedUserProperties = []*sectest.MockedFederatedUserProperties {
	{
		ExtIdpName:              "okta",
		ExtIdName:               "email",
		ExtIdValue:              "test@example.com",
	},
}