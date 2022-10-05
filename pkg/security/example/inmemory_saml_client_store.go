package example

import (
	"context"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	saml_auth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso"
	"errors"
)

type InMemorySamlClientStore struct {
	details []saml_auth.DefaultSamlClient
}

func NewInMemSpManager() samlctx.SamlClientStore {
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

func (i *InMemorySamlClientStore) GetAllSamlClient(context.Context) ([]samlctx.SamlClient, error) {
	var result []samlctx.SamlClient
	for _, v := range i.details {
		result = append(result, v)
	}
	return result, nil
}

func (i *InMemorySamlClientStore) GetSamlClientByEntityId(ctx context.Context, entityId string) (samlctx.SamlClient, error) {
	for _, detail := range i.details {
		if detail.EntityId == entityId {
			return detail, nil
		}
	}
	return saml_auth.DefaultSamlClient{}, errors.New("not found")
}