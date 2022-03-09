package example

import (
	"context"
	saml_auth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso"
	saml_auth_ctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/saml_sso_ctx"
	"errors"
)

type InMemorySamlClientStore struct {
	details []saml_auth.DefaultSamlClient
}

func NewInMemSpManager() saml_auth_ctx.SamlClientStore {
	return &InMemorySamlClientStore{
		details: []saml_auth.DefaultSamlClient{
			saml_auth.DefaultSamlClient{
				SamlSpDetails: saml_auth.SamlSpDetails{
					EntityId: "18.205.202.124",
					MetadataSource: "http://localhost:9090/metadata",
					SkipAssertionEncryption: true,
					SkipAuthRequestSignatureVerification: false,
				},
			},
			saml_auth.DefaultSamlClient{
				SamlSpDetails: saml_auth.SamlSpDetails{
					EntityId: "http://localhost:8000/saml/metadata",
					MetadataSource: "http://localhost:8000/saml/metadata",
					SkipAssertionEncryption: true,
					SkipAuthRequestSignatureVerification: false,
				},
			},
		},
	}
}

func (i *InMemorySamlClientStore) GetAllSamlClient(context.Context) ([]saml_auth_ctx.SamlClient, error) {
	var result []saml_auth_ctx.SamlClient
	for _, v := range i.details {
		result = append(result, v)
	}
	return result, nil
}

func (i *InMemorySamlClientStore) GetSamlClientByEntityId(ctx context.Context, entityId string) (saml_auth_ctx.SamlClient, error) {
	for _, detail := range i.details {
		if detail.EntityId == entityId {
			return detail, nil
		}
	}
	return saml_auth.DefaultSamlClient{}, errors.New("not found")
}