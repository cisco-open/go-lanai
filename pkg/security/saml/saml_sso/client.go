package saml_auth

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"

type SamlClient interface {
	GetEntityId() string
	GetMetadataSource() string
	ShouldSkipAssertionEncryption() bool
	ShouldSkipAuthRequestSignatureVerification() bool
	GetTenantRestrictions() utils.StringSet
}

type DefaultSamlClient struct {
	SamlSpDetails
	TenantRestrictions utils.StringSet
}

func (c DefaultSamlClient) GetEntityId() string {
	return c.EntityId
}

func (c DefaultSamlClient) GetMetadataSource() string {
	return c.MetadataSource
}

func (c DefaultSamlClient) ShouldSkipAssertionEncryption() bool {
	return c.SkipAssertionEncryption
}

func (c DefaultSamlClient) ShouldSkipAuthRequestSignatureVerification() bool {
	return c.SkipAuthRequestSignatureVerification
}

func (c DefaultSamlClient) GetTenantRestrictions() utils.StringSet {
	return c.TenantRestrictions
}


type SamlSpDetails struct {
	EntityId string
	MetadataSource string
	SkipAssertionEncryption bool
	SkipAuthRequestSignatureVerification bool

	//TODO do this for idp as well
	//metadatarequiresignature boolean,
	//metadatatrustcheck boolean,
	//metadatatrustedkeys set<text>,  //this needs an additional API to get key locations
	//securityprofile text,
}

type SamlClientStore interface {
	GetAllSamlClient() []SamlClient
	GetSamlClientById(id string) (SamlClient, error)
}