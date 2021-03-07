package samlidp

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"

type SamlIdpDetails struct {
	EntityId         string
	Domain           string
	MetadataLocation string
	ExternalIdName   string
	ExternalIdpName  string
	//TODO: option to require metadata to have signature, option to verify metadata signature
	// this is optional because both Spring and Okta's metadata are not signed
}


type SamlIdpOptions func(opt *SamlIdpDetails)

// SamlIdentityProvider implements idp.IdentityProvider, idp.AuthenticationFlowAware and samllogin.SamlIdentityProvider
type SamlIdentityProvider struct {
	SamlIdpDetails
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