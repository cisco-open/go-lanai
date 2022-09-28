package testdata

import (
	"context"
	saml_auth_ctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/saml_sso_ctx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/xml"
	"fmt"
	"github.com/crewjam/saml"
)

type TestSamlClientStore map[string]saml_auth_ctx.SamlClient

func NewTestSamlClientStore(sps ...*saml.ServiceProvider) saml_auth_ctx.SamlClientStore {
	ret := TestSamlClientStore{}
	for _, sp := range sps {
		if client := NewTestSamlClient(sp); client != nil {
			ret[client.EntityID] = client
		}
	}
	return ret
}

func (s TestSamlClientStore) GetAllSamlClient(_ context.Context) ([]saml_auth_ctx.SamlClient, error) {
	clients := make([]saml_auth_ctx.SamlClient, 0, len(s))
	for _, c := range s {
		clients = append(clients, c)
	}
	return clients, nil
}

func (s TestSamlClientStore) GetSamlClientByEntityId(_ context.Context, entityId string) (saml_auth_ctx.SamlClient, error) {
	if c, ok := s[entityId]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("cannot find SAML client with entityID [%s]", entityId)
}

type TestSamlClient struct {
	saml.ServiceProvider
	MetadataSource string
}

func NewTestSamlClient(sp *saml.ServiceProvider) *TestSamlClient {
	metadata := sp.Metadata()
	data, e := xml.Marshal(metadata)
	if e != nil {
		return nil
	}
	return &TestSamlClient{
		ServiceProvider: *sp,
		MetadataSource: string(data),
	}
}

func (c *TestSamlClient) GetEntityId() string {
	return c.EntityID
}

func (c *TestSamlClient) GetMetadataSource() string {
	return c.MetadataSource
}

func (c *TestSamlClient) ShouldSkipAssertionEncryption() bool {
	return false
}

func (c *TestSamlClient) ShouldSkipAuthRequestSignatureVerification() bool {
	return false
}

func (c *TestSamlClient) GetTenantRestrictions() utils.StringSet {
	return utils.NewStringSet()
}

func (c *TestSamlClient) GetTenantRestrictionType() string {
	return "all"
}

func (c *TestSamlClient) ShouldMetadataRequireSignature() bool {
	return false
}

func (c *TestSamlClient) ShouldMetadataTrustCheck() bool {
	return false
}

func (c *TestSamlClient) GetMetadataTrustedKeys() []string {
	return nil
}
