package example

import (
	"context"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	samlidp "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/idp"
	"errors"
)

type InMemorySamlClientStore struct {
	details []samlidp.DefaultSamlClient
}

func NewInMemSpManager() samlctx.SamlClientStore {
	return &InMemorySamlClientStore{
		details: []samlidp.DefaultSamlClient{
			//TODO: populate your SAML SP details here
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
	return samlidp.DefaultSamlClient{}, errors.New("not found")
}
