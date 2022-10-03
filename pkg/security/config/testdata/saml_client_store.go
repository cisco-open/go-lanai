package testdata

import (
	saml_auth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso"
	saml_auth_ctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/saml_sso_ctx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samlssotest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"fmt"
)

func NewMockedSamlClientStore(props ...*sectest.MockedClientProperties) *samlssotest.MockSamlClientStore {
	samlClients := make([]saml_auth_ctx.SamlClient, len(props))
	for i, v := range props {
		samlClients[i] = newMockedSamlClient(v)
	}
	return samlssotest.NewMockedSamlClientStore(samlClients...).(*samlssotest.MockSamlClientStore)
}

func newMockedSamlClient(props *sectest.MockedClientProperties) *saml_auth.DefaultSamlClient {
	entityId := fmt.Sprintf("http://%s/saml/metadata", props.ClientID)
	metadataSource := fmt.Sprintf("testdata/%s-saml-metadata.xml")
	return &saml_auth.DefaultSamlClient {
		SamlSpDetails: saml_auth.SamlSpDetails{
			EntityId:                             entityId,
			MetadataSource:                       metadataSource,
			SkipAssertionEncryption:              false,
			SkipAuthRequestSignatureVerification: false,
		},
		TenantRestrictions: utils.NewStringSet(),
		TenantRestrictionType: "any",
	}
}