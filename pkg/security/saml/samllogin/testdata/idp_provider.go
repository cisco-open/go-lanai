package testdata

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"

type TestIdpProvider struct {
	domain string
	metadataLocation string
	externalIdpName string
	externalIdName string
	entityId string
	metadataRequireSignature bool
	metadataTrustCheck bool
	metadataTrustedKeys []string
}

func (i TestIdpProvider) GetAutoCreateUserDetails() security.AutoCreateUserDetails {
	panic("implement me")
}

func (i TestIdpProvider) ShouldMetadataRequireSignature() bool {
	return i.metadataRequireSignature
}

func (i TestIdpProvider) ShouldMetadataTrustCheck() bool {
	return i.metadataTrustCheck
}

func (i TestIdpProvider) GetMetadataTrustedKeys() []string {
	return i.metadataTrustedKeys
}

func (i TestIdpProvider) Domain() string {
	return i.domain
}

func (i TestIdpProvider) EntityId() string {
	return i.entityId
}

func (i TestIdpProvider) MetadataLocation() string {
	return i.metadataLocation
}

func (i TestIdpProvider) ExternalIdName() string {
	return i.externalIdName
}

func (i TestIdpProvider) ExternalIdpName() string {
	return i.externalIdpName
}

