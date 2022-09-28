package testdata

import (
	"crypto/rsa"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"fmt"
	"github.com/crewjam/saml"
	dsig "github.com/russellhaering/goxmldsig"
	"net/url"
)

const (
	TestIdpSsoPath = "/authorize"
	TestIdpSloPath = "/logout"

	TestSPEntityID    = `http://samlsp.msx.com/samlsp/saml/metadata`
	TestSPCertFile    = `testdata/saml_test_sp.cert`
	TestSPPrivKeyFile = `testdata/saml_test_sp.key`
	TestSPAcsURL      = `http://samlsp.msx.com/samlsp/saml/acs`
	TestSPSloURL      = `http://samlsp.msx.com/samlsp/saml/slo`
	TestIDPCertFile   = `testdata/saml_test.cert`
)

var (
	TestIdpURL, _ = TestIssuer.BuildUrl()
	TestIdpSsoURL = TestIdpURL.ResolveReference(&url.URL{Path: fmt.Sprintf("%s%s?grant_type=%s", TestIdpURL.Path, TestIdpSsoPath, oauth2.GrantTypeSamlSSO)})
	TestIdpSloURL = TestIdpURL.ResolveReference(&url.URL{Path: fmt.Sprintf("%s%s", TestIdpURL.Path, TestIdpSloPath)})
)

func NewTestSP() *saml.ServiceProvider {
	var e error
	var spCerts, idpCerts []*x509.Certificate
	var privKey *rsa.PrivateKey
	var acsUrl, sloUrl *url.URL

	if spCerts, e = cryptoutils.LoadCert(TestSPCertFile); e != nil {
		panic(e)
	}
	if idpCerts, e = cryptoutils.LoadCert(TestIDPCertFile); e != nil {
		panic(e)
	}
	if privKey, e = cryptoutils.LoadPrivateKey(TestSPPrivKeyFile, ""); e != nil {
		panic(e)
	}
	if acsUrl, e = url.Parse(TestSPAcsURL); e != nil {
		panic(e)
	}
	if sloUrl, e = url.Parse(TestSPSloURL); e != nil {
		panic(e)
	}

	testIDP := &saml.IdentityProvider{
		Certificate: idpCerts[0],
		MetadataURL: *TestIdpURL,
		SSOURL:      *TestIdpSsoURL,
		LogoutURL:   *TestIdpSloURL,
	}
	sp := saml.ServiceProvider{
		EntityID:          TestSPEntityID,
		Key:               privKey,
		Certificate:       spCerts[0],
		AcsURL:            *acsUrl,
		SloURL:            *sloUrl,
		SignatureMethod:   dsig.RSASHA256SignatureMethod,
		AllowIDPInitiated: true,
		AuthnNameIDFormat: saml.UnspecifiedNameIDFormat,
		LogoutBindings:    []string{saml.HTTPPostBinding},
		IDPMetadata:       testIDP.Metadata(),
	}
	return &sp
}
