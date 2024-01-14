package service

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"fmt"
)

var (
	globalLocalhostIdp = passwdidp.NewIdentityProvider(func(opt *passwdidp.PasswdIdpDetails) {
		opt.Domain = "localhost"
	})
)

// InMemoryIdpManager implements idp.IdentityProviderManager, samllogin.SamlIdentityProviderManager
type InMemoryIdpManager struct{}

// idp.IdentityProviderManager
func (i *InMemoryIdpManager) GetIdentityProvidersWithFlow(ctx context.Context, flow idp.AuthenticationFlow) []idp.IdentityProvider {
	switch flow {
	case idp.ExternalIdpSAML:
		return []idp.IdentityProvider{}
	case idp.InternalIdpForm:
		return []idp.IdentityProvider{
			globalLocalhostIdp,
		}
	}
	return []idp.IdentityProvider{}
}

// idp.IdentityProviderManager
func (i *InMemoryIdpManager) GetIdentityProviderByDomain(ctx context.Context, domain string) (idp.IdentityProvider, error) {
	switch {
	case domain == globalLocalhostIdp.Domain():
		return globalLocalhostIdp, nil
	}
	return nil, fmt.Errorf("cannot find IDP with domain %s", domain)
}

// samllogin.SamlIdentityProviderManager
func (i *InMemoryIdpManager) GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error) {
	return nil, fmt.Errorf("cannot find IDP with entity ID %s", entityId)
}

func NewInMemoryIdpManager() idp.IdentityProviderManager {
	return &InMemoryIdpManager{}
}
