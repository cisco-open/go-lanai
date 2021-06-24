package example

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/samlidp"
	"fmt"
	"strings"
)

var (
	globalSamlIdp = samlidp.NewIdentityProvider(func(opt *samlidp.SamlIdpDetails) {
		opt.Domain = "saml.vms.com"
		opt.MetadataLocation = "https://dev-940621.oktapreview.com/app/exkwj65c2kC1vwtYi0h7/sso/saml/metadata"
		opt.ExternalIdpName = "okta"
		opt.ExternalIdName = "email"
		opt.EntityId = "http://www.okta.com/exkwj65c2kC1vwtYi0h7"
	})

	globalPasswdIdp = passwdidp.NewIdentityProvider(func(opt *passwdidp.PasswdIdpDetails) {
		opt.Domain = "internal.vms.com"
	})

	globalLocalhostIdp = passwdidp.NewIdentityProvider(func(opt *passwdidp.PasswdIdpDetails) {
		opt.Domain = "localhost"
	})
)

// InMemoryIdpManager implements idp.IdentityProviderManager, samllogin.SamlIdentityProviderManager
type InMemoryIdpManager struct {}

// idp.IdentityProviderManager
func (i *InMemoryIdpManager) GetIdentityProvidersWithFlow(ctx context.Context, flow idp.AuthenticationFlow) []idp.IdentityProvider {
	switch flow {
	case idp.ExternalIdpSAML:
		return []idp.IdentityProvider{
			globalSamlIdp,
		}
	case idp.InternalIdpForm:
		return []idp.IdentityProvider{
			globalPasswdIdp, globalLocalhostIdp,
		}
	}
	return []idp.IdentityProvider{}
}

// idp.IdentityProviderManager
func (i *InMemoryIdpManager) GetIdentityProviderByDomain(ctx context.Context, domain string) (idp.IdentityProvider, error) {
	switch {
	case domain == globalSamlIdp.Domain():
		return globalSamlIdp, nil
	case strings.HasSuffix(domain, globalPasswdIdp.Domain()) && !strings.HasPrefix(domain, "."):
		return globalPasswdIdp, nil
	case domain == globalLocalhostIdp.Domain():
		return globalLocalhostIdp, nil
	}
	return nil, fmt.Errorf("cannot find IDP with domain %s", domain)
}

// samllogin.SamlIdentityProviderManager
func (i *InMemoryIdpManager) GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error) {
	if entityId == globalSamlIdp.EntityId() {
		return globalSamlIdp, nil
	}
	return nil, fmt.Errorf("cannot find IDP with entity ID %s", entityId)
}

func NewInMemoryIdpManager() idp.IdentityProviderManager {
	return &InMemoryIdpManager{}
}
