package example

import (
	saml_auth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml_sso"
	"errors"
)

type InMemorySamlClientStore struct {
	details []saml_auth.DefaultSamlClient
}

func NewInMemSpManager() saml_auth.SamlClientStore {
	return &InMemorySamlClientStore{
		details: []saml_auth.DefaultSamlClient{
			saml_auth.DefaultSamlClient{
				SamlSpDetails: saml_auth.SamlSpDetails {
					EntityId: "18.205.202.124",
					MetadataSource: "http://localhost:9090/metadata",
					SkipAssertionEncryption: true,
					SkipAuthRequestSignatureVerification: false,
				},
			},
		},
	}
}

func (i *InMemorySamlClientStore) GetAllSamlClient() []saml_auth.SamlClient {
	var result []saml_auth.SamlClient
	for _, v := range i.details {
		result = append(result, v)
	}
	return result
}

func (i *InMemorySamlClientStore) GetSamlClientById(id string) (saml_auth.SamlClient, error) {
	if i.details[0].EntityId == id {
		return i.details[0], nil
	} else {
		return saml_auth.DefaultSamlClient{}, errors.New("not found")
	}
}

