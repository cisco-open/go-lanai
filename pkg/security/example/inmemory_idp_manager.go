package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/samllogin"
)

type InMemoryIdpManager struct {}

func (i *InMemoryIdpManager) GetAllIdentityProvider() []idp.IdentityProviderDetails {
	return []idp.IdentityProviderDetails{
		samllogin.SamlIdpDetails{
			Domain:           "saml.vms.com",
			MetadataLocation: "https://dev-940621.oktapreview.com/app/exkwj65c2kC1vwtYi0h7/sso/saml/metadata",
			ExternalIdpName: "okta",
			ExternalIdName: "email",
			EntityId: "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
		}}
}

func (i *InMemoryIdpManager) GetIdentityProviderByEntityId(entityId string) (idp.IdentityProviderDetails, error) {
	return samllogin.SamlIdpDetails{
		Domain:           "saml.vms.com",
		MetadataLocation: "https://dev-940621.oktapreview.com/app/exkwj65c2kC1vwtYi0h7/sso/saml/metadata",
		ExternalIdpName: "okta",
		ExternalIdName: "email",
		EntityId: "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
	}, nil
}

func NewInMemoryIdpManager() idp.IdentityProviderManager {
	return &InMemoryIdpManager{}
}
