package example

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/samllogin"

type InMemoryIdpManager struct {}

func (i *InMemoryIdpManager) GetAllIdentityProvider() []samllogin.IdentityProviderDetails {
	return []samllogin.IdentityProviderDetails{
		samllogin.IdentityProviderDetails{
			Domain:           "saml.vms.com",
			MetadataLocation: "https://dev-940621.oktapreview.com/app/exkwj65c2kC1vwtYi0h7/sso/saml/metadata",
			ExternalIdpName: "okta",
			ExternalIdName: "email",
			EntityId: "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
		}}
}

func (i *InMemoryIdpManager) GetIdentityProviderByEntityId(entityId string) (samllogin.IdentityProviderDetails, error) {
	return samllogin.IdentityProviderDetails{
		Domain:           "saml.vms.com",
		MetadataLocation: "https://dev-940621.oktapreview.com/app/exkwj65c2kC1vwtYi0h7/sso/saml/metadata",
		ExternalIdpName: "okta",
		ExternalIdName: "email",
		EntityId: "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
	}, nil
}

func NewInMemoryIdpManager() samllogin.IdentityProviderManager {
	return &InMemoryIdpManager{}
}
