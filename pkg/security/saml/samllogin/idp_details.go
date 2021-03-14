package samllogin

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"

type SamlIdentityProvider interface {
	idp.IdentityProvider
	EntityId() string
	MetadataLocation() string
	ExternalIdName() string
	ExternalIdpName() string
}

type SamlIdentityProviderManager interface {
	GetIdentityProviderByEntityId(entityId string) (idp.IdentityProvider, error)
}
