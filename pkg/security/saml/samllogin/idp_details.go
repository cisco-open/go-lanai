package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
)

type SamlIdentityProvider interface {
	idp.IdentityProvider
	EntityId() string
	MetadataLocation() string
	ExternalIdName() string
	ExternalIdpName() string
	ShouldMetadataRequireSignature() bool
	ShouldMetadataTrustCheck() bool
	GetMetadataTrustedKeys() []string
}

type SamlIdentityProviderManager interface {
	GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error)
}

