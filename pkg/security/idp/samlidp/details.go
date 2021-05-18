package samlidp

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"

type SamlIdpDetails struct {
	EntityId         string
	Domain           string
	MetadataLocation string
	ExternalIdName   string
	ExternalIdpName  string
	MetadataRequireSignature bool
	MetadataTrustCheck bool
	MetadataTrustedKeys []string
}


type SamlIdpOptions func(opt *SamlIdpDetails)

// SamlIdentityProvider implements idp.IdentityProvider, idp.AuthenticationFlowAware and samllogin.SamlIdentityProvider
type SamlIdentityProvider struct {
	SamlIdpDetails
}

func (s SamlIdentityProvider) ShouldMetadataRequireSignature() bool {
	return s.MetadataRequireSignature
}

func (s SamlIdentityProvider) ShouldMetadataTrustCheck() bool {
	return s.MetadataTrustCheck
}

func (s SamlIdentityProvider) GetMetadataTrustedKeys() []string {
	return s.MetadataTrustedKeys
}

func NewIdentityProvider(opts ...SamlIdpOptions) *SamlIdentityProvider {
	opt := SamlIdpDetails{}
	for _, f := range opts {
		f(&opt)
	}
	return &SamlIdentityProvider{
		SamlIdpDetails: opt,
	}
}

func (s SamlIdentityProvider) AuthenticationFlow() idp.AuthenticationFlow {
	return idp.ExternalIdpSAML
}

func (s SamlIdentityProvider) Domain() string {
	return s.SamlIdpDetails.Domain
}

func (s SamlIdentityProvider) EntityId() string {
	return s.SamlIdpDetails.EntityId
}

func (s SamlIdentityProvider) MetadataLocation() string {
	return s.SamlIdpDetails.MetadataLocation
}

func (s SamlIdentityProvider) ExternalIdName() string {
	return s.SamlIdpDetails.ExternalIdName
}

func (s SamlIdentityProvider) ExternalIdpName() string {
	return s.SamlIdpDetails.ExternalIdpName
}