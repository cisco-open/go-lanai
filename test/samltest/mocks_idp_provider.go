package samltest

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"

type ExtSamlMetadata struct {
	EntityId         string
	Domain           string
	Source           string
	Name             string
	IdName           string
	RequireSignature bool
	TrustCheck       bool
	TrustedKeys      []string
}

func NewMockedIdpProvider(opts ...IDPMockOptions) *MockedIdpProvider {
	defaultEntityID, _ := DefaultIssuer.BuildUrl()
	opt := IDPMockOption{
		Properties: IDPProperties{
			ProviderProperties: ProviderProperties{
				EntityID: defaultEntityID.String(),
			},
			SSOPath: "/sso",
			SLOPath: "/slo",
		},
	}
	for _, fn := range opts {
		fn(&opt)
	}
	return &MockedIdpProvider{ExtSamlMetadata{
		EntityId:         opt.Properties.EntityID,
		Domain:           opt.Properties.Domain,
		Source:           opt.Properties.MetadataSource,
		Name:             opt.Properties.Name,
		IdName:           opt.Properties.IdName,
	}}
}

type MockedIdpProvider struct {
	ExtSamlMetadata
}

func (i MockedIdpProvider) Domain() string {
	return i.ExtSamlMetadata.Domain
}

func (i MockedIdpProvider) GetAutoCreateUserDetails() security.AutoCreateUserDetails {
	return nil
}

func (i MockedIdpProvider) ShouldMetadataRequireSignature() bool {
	return i.ExtSamlMetadata.RequireSignature
}

func (i MockedIdpProvider) ShouldMetadataTrustCheck() bool {
	return i.ExtSamlMetadata.TrustCheck
}

func (i MockedIdpProvider) GetMetadataTrustedKeys() []string {
	return i.ExtSamlMetadata.TrustedKeys
}

func (i MockedIdpProvider) EntityId() string {
	return i.ExtSamlMetadata.EntityId
}

func (i MockedIdpProvider) MetadataLocation() string {
	return i.ExtSamlMetadata.Source
}

func (i MockedIdpProvider) ExternalIdName() string {
	return i.ExtSamlMetadata.IdName
}

func (i MockedIdpProvider) ExternalIdpName() string {
	return i.ExtSamlMetadata.Name
}
