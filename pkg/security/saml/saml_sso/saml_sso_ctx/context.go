// TODO move interface to "saml" package

package saml_auth_ctx

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

type SamlClient interface {
	GetEntityId() string
	GetMetadataSource() string
	ShouldSkipAssertionEncryption() bool
	ShouldSkipAuthRequestSignatureVerification() bool
	GetTenantRestrictions() utils.StringSet
	GetTenantRestrictionType() string

	ShouldMetadataRequireSignature() bool
	ShouldMetadataTrustCheck() bool
	GetMetadataTrustedKeys() []string
}

type SamlClientStore interface {
	GetAllSamlClient(ctx context.Context) ([]SamlClient, error)
	GetSamlClientByEntityId(ctx context.Context, entityId string) (SamlClient, error)
}