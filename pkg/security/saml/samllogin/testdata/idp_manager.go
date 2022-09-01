package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"errors"
	"sort"
)

type TestIdpManager struct {
	idpDetails []TestIdpProvider
}

func NewTestIdpManager(idps ...TestIdpProvider) *TestIdpManager {
	details := []TestIdpProvider{
		{
			domain:           "saml.vms.com",
			metadataLocation: "testdata/okta_login_test_metadata.xml",
			externalIdpName:  "okta",
			externalIdName:   "email",
			entityId:         "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
		},
		{
			domain:           "saml-alt.vms.com",
			metadataLocation: "testdata/okta_logout_test_metadata.xml",
			externalIdpName:  "okta",
			externalIdName:   "email",
			entityId:         "http://www.okta.com/exk668ha29xaI4in25d7",
		},
	}
	for _, v := range idps {
		details = append(details, v)
	}
	return &TestIdpManager{
		idpDetails: details,
	}
}

func (t TestIdpManager) GetIdentityProvidersWithFlow(context.Context, idp.AuthenticationFlow) (ret []idp.IdentityProvider) {
	ret = make([]idp.IdentityProvider, len(t.idpDetails))
	for i, v := range t.idpDetails {
		ret[i] = v
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].Domain() < ret[j].Domain()
	})
	return
}

func (t TestIdpManager) GetIdentityProviderByEntityId(_ context.Context, entityId string) (idp.IdentityProvider, error) {
	for _, v := range t.idpDetails {
		if entityId == v.entityId {
			return v, nil
		}
	}
	return nil, errors.New("not found")
}

func (t TestIdpManager) GetIdentityProviderByDomain(_ context.Context, domain string) (idp.IdentityProvider, error) {
	for _, v := range t.idpDetails {
		if domain == v.domain {
			return v, nil
		}
	}
	return nil, errors.New("not found")
}
