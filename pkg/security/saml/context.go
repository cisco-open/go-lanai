package samlctx

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

/********************
	For IDP
 ********************/

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

/********************
	For SP
 ********************/

type SamlIdentityProvider interface {
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

type SamlIdentityProviderManager interface {
	GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error)
}

// SamlBindingManager is an additional interface that SamlIdentityProviderManager could implement.
type SamlBindingManager interface {
	// PreferredBindings returns supported bindings in order of preference.
	// possible values are
	// - saml.HTTPRedirectBinding
	// - saml.HTTPPostBinding
	// - saml.HTTPArtifactBinding
	// - saml.SOAPBinding
	// Note that this is not list of supported bindings. Supported bindings are determined by IDP and SP
	PreferredBindings() []string
}