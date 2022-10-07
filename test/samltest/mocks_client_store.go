package samltest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	saml_auth_ctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso/saml_sso_ctx"
	"errors"
	"github.com/crewjam/saml"
)

type ClientStoreMockOptions func(opt *ClientStoreMockOption)
type ClientStoreMockOption struct {
	Clients []saml_auth_ctx.SamlClient
	SPs []*saml.ServiceProvider
	ClientsProperties map[string]MockedClientProperties
}

// ClientsWithPropertiesPrefix returns a ClientStoreMockOptions that bind a map of properties from application config with given prefix
func ClientsWithPropertiesPrefix(appCfg bootstrap.ApplicationConfig, prefix string) ClientStoreMockOptions {
	return func(opt *ClientStoreMockOption) {
		if e := appCfg.Bind(&opt.ClientsProperties, prefix); e != nil {
			panic(e)
		}
	}
}

// ClientsWithSPs returns a ClientStoreMockOptions that convert given SPs to Clients
func ClientsWithSPs(sps...*saml.ServiceProvider) ClientStoreMockOptions {
	return func(opt *ClientStoreMockOption) {
		opt.SPs = sps
	}
}

type MockSamlClientStore struct {
	details []saml_auth_ctx.SamlClient
}

func NewMockedClientStore(opts...ClientStoreMockOptions) *MockSamlClientStore {
	opt := ClientStoreMockOption {}
	for _, fn := range opts {
		fn(&opt)
	}

	var details []saml_auth_ctx.SamlClient
	switch {
	case len(opt.Clients) > 0:
		details = opt.Clients
	case len(opt.SPs) > 0:
		for _, sp := range opt.SPs {
			v := NewMockedSamlClient(func(opt *MockedClientOption) {
				opt.SP = sp
			})
			details = append(details, v)
		}
	default:
		for _, props := range opt.ClientsProperties {
			v := NewMockedSamlClient(func(opt *MockedClientOption) {
				opt.Properties = props
			})
			details = append(details, v)
		}
	}

	return &MockSamlClientStore{details: details}
}

func (t *MockSamlClientStore) GetAllSamlClient(_ context.Context) ([]saml_auth_ctx.SamlClient, error) {
	var result []saml_auth_ctx.SamlClient
	for _, v := range t.details {
		result = append(result, v)
	}
	return result, nil
}

func (t *MockSamlClientStore) GetSamlClientByEntityId(_ context.Context, id string) (saml_auth_ctx.SamlClient, error) {
	for _, detail := range t.details {
		if detail.GetEntityId() == id {
			return detail, nil
		}
	}
	return nil, errors.New("not found")
}

