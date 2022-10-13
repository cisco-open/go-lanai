package samltest

import (
	"crypto/rsa"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"github.com/crewjam/saml"
	dsig "github.com/russellhaering/goxmldsig"
	"net/url"
)

type IDPMockOptions func(opt *IDPMockOption)
type IDPMockOption struct {
	Properties IDPProperties
}

// IDPWithPropertiesPrefix returns a IDP mock option that bind properties from application config and with given prefix
func IDPWithPropertiesPrefix(appCfg bootstrap.ApplicationConfig, prefix string) IDPMockOptions {
	return func(opt *IDPMockOption) {
		if e := appCfg.Bind(&opt.Properties, prefix); e != nil {
			panic(e)
		}
	}
}

// MustNewMockedIDP similar to NewMockedIDP, panic instead of returning error
func MustNewMockedIDP(opts ...IDPMockOptions) *saml.IdentityProvider {
	sp, e := NewMockedIDP(opts...)
	if e != nil {
		panic(e)
	}
	return sp
}

// NewMockedIDP create a mocked IDP with given IDPMockOptions.
// Returns error if any mocked value are incorrect. e.g. file not exists
func NewMockedIDP(opts ...IDPMockOptions) (*saml.IdentityProvider, error) {
	defaultEntityID, _ := DefaultIssuer.BuildUrl()
	opt := IDPMockOption{
		Properties: IDPProperties{
			ProviderProperties: ProviderProperties{
				EntityID: defaultEntityID.String(),
			},
			SSOPath: "/sso",
			SLOPath: "/slo",
		},
	}
	for _, fn := range opts {
		fn(&opt)
	}

	var e error
	var certs []*x509.Certificate
	var privKey *rsa.PrivateKey
	var metaUrl, ssoUrl, sloUrl *url.URL
	if certs, e = cryptoutils.LoadCert(opt.Properties.CertsSource); e != nil && len(opt.Properties.CertsSource) != 0 {
		return nil, e
	}
	if privKey, e = cryptoutils.LoadPrivateKey(opt.Properties.PrivateKeySource, ""); e != nil && len(opt.Properties.PrivateKeySource) != 0 {
		return nil, e
	}
	if metaUrl, e = resolveAbsUrl(opt.Properties.EntityID, opt.Properties.EntityID); e != nil && len(opt.Properties.EntityID) != 0 {
		return nil, e
	}
	if ssoUrl, e = resolveAbsUrl(opt.Properties.EntityID, opt.Properties.SSOPath); e != nil && len(opt.Properties.SSOPath) != 0 {
		return nil, e
	}
	if sloUrl, e = resolveAbsUrl(opt.Properties.EntityID, opt.Properties.SLOPath); e != nil && len(opt.Properties.SLOPath) != 0 {
		return nil, e
	}

	return &saml.IdentityProvider{
		Key:             privKey,
		Certificate:     certs[0],
		MetadataURL:     *metaUrl,
		SSOURL:          *ssoUrl,
		LogoutURL:       *sloUrl,
		SignatureMethod: dsig.RSASHA256SignatureMethod,
	}, nil
}
