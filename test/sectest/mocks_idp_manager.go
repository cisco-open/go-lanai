package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"fmt"
)

type MockedPasswdIdentityProvider struct {
	domain string
}

func NewMockedPasswdIdentityProvider(domain string) *MockedPasswdIdentityProvider {
	return &MockedPasswdIdentityProvider{
		domain: domain,
	}
}

func (s MockedPasswdIdentityProvider) AuthenticationFlow() idp.AuthenticationFlow {
	return idp.InternalIdpForm
}

func (s MockedPasswdIdentityProvider) Domain() string {
	return s.domain
}

type MockedIDPManager struct {
	idpPasswd idp.IdentityProvider
}

type IdpManagerMockOptions func(opt *IdpManagerMockOption)
type IdpManagerMockOption struct {
	PasswdIDPDomain string
}

func NewMockedIDPManager(opts...IdpManagerMockOptions) *MockedIDPManager {
	opt := IdpManagerMockOption{}
	for _, fn := range opts {
		fn(&opt)
	}
	return &MockedIDPManager{
		idpPasswd: NewMockedPasswdIdentityProvider(opt.PasswdIDPDomain),
	}
}

func (m *MockedIDPManager) GetIdentityProvidersWithFlow(ctx context.Context, flow idp.AuthenticationFlow) []idp.IdentityProvider {
	switch flow {
	case idp.InternalIdpForm:
		return []idp.IdentityProvider{m.idpPasswd}
	default:
		return []idp.IdentityProvider{}
	}
}

func (m *MockedIDPManager) GetIdentityProviderByDomain(ctx context.Context, domain string) (idp.IdentityProvider, error) {
	switch domain {
	case m.idpPasswd.Domain():
		return m.idpPasswd, nil
	}
	return nil, fmt.Errorf("cannot find IDP for domain [%s]", domain)
}
