package samltest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/xml"
	"github.com/crewjam/saml"
)

type MockedClientOptions func(opt *MockedClientOption)
type MockedClientOption struct {
	Properties MockedClientProperties
	SP         *saml.ServiceProvider
}

type MockedSamlClient struct {
	EntityId                             string
	MetadataSource                       string
	SkipAssertionEncryption              bool
	SkipAuthRequestSignatureVerification bool
	MetadataRequireSignature             bool
	MetadataTrustCheck                   bool
	MetadataTrustedKeys                  []string
	TenantRestrictions                   utils.StringSet
	TenantRestrictionType                string
}

func NewMockedSamlClient(opts ...MockedClientOptions) *MockedSamlClient {
	opt := MockedClientOption{}
	for _, fn := range opts {
		fn(&opt)
	}

	if opt.SP != nil {
		metadata := opt.SP.Metadata()
		data, e := xml.Marshal(metadata)
		if e != nil {
			return nil
		}
		return &MockedSamlClient{
			EntityId:              opt.SP.EntityID,
			MetadataSource:        string(data),
			TenantRestrictions:    utils.NewStringSet(),
			TenantRestrictionType: "all",
		}
	}

	return &MockedSamlClient{
		EntityId:                             opt.Properties.EntityID,
		MetadataSource:                       opt.Properties.MetadataSource,
		SkipAssertionEncryption:              opt.Properties.SkipEncryption,
		SkipAuthRequestSignatureVerification: opt.Properties.SkipSignatureVerification,
		TenantRestrictions:                   utils.NewStringSet(opt.Properties.TenantRestriction...),
		TenantRestrictionType:                opt.Properties.TenantRestrictionType,
	}
}

func (c MockedSamlClient) ShouldMetadataRequireSignature() bool {
	return c.MetadataRequireSignature
}

func (c MockedSamlClient) ShouldMetadataTrustCheck() bool {
	return c.MetadataTrustCheck
}

func (c MockedSamlClient) GetMetadataTrustedKeys() []string {
	return c.MetadataTrustedKeys
}

func (c MockedSamlClient) GetEntityId() string {
	return c.EntityId
}

func (c MockedSamlClient) GetMetadataSource() string {
	return c.MetadataSource
}

func (c MockedSamlClient) ShouldSkipAssertionEncryption() bool {
	return c.SkipAssertionEncryption
}

func (c MockedSamlClient) ShouldSkipAuthRequestSignatureVerification() bool {
	return c.SkipAuthRequestSignatureVerification
}

func (c MockedSamlClient) GetTenantRestrictions() utils.StringSet {
	return c.TenantRestrictions
}

func (c MockedSamlClient) GetTenantRestrictionType() string {
	return c.TenantRestrictionType
}
