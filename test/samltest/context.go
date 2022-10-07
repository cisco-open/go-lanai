package samltest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
)

// TODO remove this file when interfaces are moved to "saml" package

type samlIdentityProvider interface {
	idp.IdentityProvider
	EntityId() string
	MetadataLocation() string
	ExternalIdName() string
	ExternalIdpName() string
	ShouldMetadataRequireSignature() bool
	ShouldMetadataTrustCheck() bool
	GetMetadataTrustedKeys() []string
	GetAutoCreateUserDetails() security.AutoCreateUserDetails
}

type samlIdentityProviderManager interface {
	GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error)
}
