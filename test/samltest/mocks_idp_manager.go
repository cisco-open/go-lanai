package samltest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"errors"
	"sort"
)

type MockedIdpManager struct {
	idpDetails []idp.IdentityProvider
	delegates  []idp.IdentityProviderManager
}

type IdpManagerMockOptions func(opt *IdpManagerMockOption)
type IdpManagerMockOption struct {
	IDPList       []idp.IdentityProvider
	IDPProperties map[string]IDPProperties
	Delegates     []idp.IdentityProviderManager
}

// IDPsWithPropertiesPrefix returns a IdpManagerMockOptions that bind a map of properties from application config and with given prefix
func IDPsWithPropertiesPrefix(appCfg bootstrap.ApplicationConfig, prefix string) IdpManagerMockOptions {
	return func(opt *IdpManagerMockOption) {
		if e := appCfg.Bind(&opt.IDPProperties, prefix); e != nil {
			panic(e)
		}
	}
}

// IDPsWithFallback returns a IdpManagerMockOptions that set a fallback implementation for non-SAML IDPs
func IDPsWithFallback(delegates ...idp.IdentityProviderManager) IdpManagerMockOptions {
	return func(opt *IdpManagerMockOption) {
		opt.Delegates = delegates
	}
}

// NewMockedIdpManager create a mocked samllogin.SamlIdentityProviderManager that returns SAML IDP based on given options
func NewMockedIdpManager(opts ...IdpManagerMockOptions) *MockedIdpManager {
	opt := IdpManagerMockOption{}
	for _, fn := range opts {
		fn(&opt)
	}

	var details []idp.IdentityProvider
	switch {
	case len(opt.IDPList) > 0:
		details = opt.IDPList
	default:
		for _, props := range opt.IDPProperties {
			v := NewMockedIdpProvider(func(opt *IDPMockOption) {
				opt.Properties = props
			})
			details = append(details, v)
		}
	}

	return &MockedIdpManager{
		idpDetails: details,
		delegates:  opt.Delegates,
	}
}

func (m MockedIdpManager) GetIdentityProvidersWithFlow(ctx context.Context, flow idp.AuthenticationFlow) (ret []idp.IdentityProvider) {
	ret = make([]idp.IdentityProvider, len(m.idpDetails), len(m.idpDetails) + 5)
	for i, v := range m.idpDetails {
		ret[i] = v
	}
	for _, delegate := range m.delegates {
		ret = append(ret, delegate.GetIdentityProvidersWithFlow(ctx, flow)...)
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].Domain() < ret[j].Domain()
	})
	return
}

func (m MockedIdpManager) GetIdentityProviderByEntityId(ctx context.Context, entityId string) (idp.IdentityProvider, error) {
	for _, v := range m.idpDetails {
		if samlIdp, ok := v.(samlIdentityProvider); ok && entityId == samlIdp.EntityId() {
			return v, nil
		}
	}
	for _, delegate := range m.delegates {
		samlDelegate, ok := delegate.(samlIdentityProviderManager)
		if !ok {
			continue
		}
		if v, e := samlDelegate.GetIdentityProviderByEntityId(ctx, entityId); e != nil {
			return v, nil
		}
	}
	return nil, errors.New("not found")
}

func (m MockedIdpManager) GetIdentityProviderByDomain(ctx context.Context, domain string) (idp.IdentityProvider, error) {
	for _, v := range m.idpDetails {
		if domain == v.Domain() {
			return v, nil
		}
	}
	for _, delegate := range m.delegates {
		if v, e := delegate.GetIdentityProviderByDomain(ctx, domain); e == nil {
			return v, nil
		}
	}
	return nil, errors.New("not found")
}
