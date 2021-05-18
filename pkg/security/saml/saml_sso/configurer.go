package saml_auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
	"net/url"
)

var (
	FeatureId = security.FeatureId("SamlAuthorizeEndpoint", security.FeatureOrderSamlAuthorizeEndpoint)
)

type SamlAuthorizeEndpointConfigurer struct {
	properties         samlctx.SamlProperties
	serverProperties   web.ServerProperties
	samlClientStore    SamlClientStore
	accountStore       security.AccountStore
	attributeGenerator AttributeGenerator
}

func newSamlAuthorizeEndpointConfigurer(properties samlctx.SamlProperties,
	samlClientStore SamlClientStore,
	accountStore security.AccountStore,
	attributeGenerator AttributeGenerator) *SamlAuthorizeEndpointConfigurer {

	return &SamlAuthorizeEndpointConfigurer{
		properties:         properties,
		samlClientStore:    samlClientStore,
		accountStore:       accountStore,
		attributeGenerator: attributeGenerator,
	}
}

func (c *SamlAuthorizeEndpointConfigurer) Apply(feature security.Feature, ws security.WebSecurity) (err error) {
	f := feature.(*Feature)

	opts := c.getIdentityProviderConfiguration(f)
	mw := NewSamlAuthorizeEndpointMiddleware(opts, c.samlClientStore, c.accountStore, c.attributeGenerator)

	ws.
		Add(middleware.NewBuilder("Saml Service Provider Refresh").
			ApplyTo(matcher.RouteWithPattern(f.ssoLocation.Path, http.MethodGet, http.MethodPost)).
			Order(security.	MWOrderSAMLMetadataRefresh).
			Use(mw.RefreshMetadataHandler(f.ssoCondition))).
		Add(middleware.NewBuilder("Saml SSO").
			ApplyTo(matcher.RouteWithPattern(f.ssoLocation.Path, http.MethodGet, http.MethodPost)).
			Order(security.MWOrderSamlAuthEndpoints).
			Use(mw.AuthorizeHandlerFunc(f.ssoCondition)))

	ws.Add(mapping.Get(f.ssoLocation.Path).HandlerFunc(security.NoopHandlerFunc()))
	ws.Add(mapping.Post(f.ssoLocation.Path).HandlerFunc(security.NoopHandlerFunc()))

	//metadata is an actual endpoint
	ws.Add(mapping.Get(f.metadataPath).
		HandlerFunc(mw.MetadataHandlerFunc()).
		Name("saml metadata"))

	// configure error handling
	errorhandling.Configure(ws).
		AdditionalErrorHandler(NewSamlErrorHandler())
	return nil
}

func (c *SamlAuthorizeEndpointConfigurer) getIdentityProviderConfiguration(f *Feature) Options {
	cert, err := cryptoutils.LoadCert(c.properties.CertificateFile)
	if err != nil {
		panic(security.NewInternalError("cannot load certificate from file", err))
	}
	if len(cert) > 1 {
		logger.Warnf("multiple certificate found, using first one")
	}
	key, err := cryptoutils.LoadPrivateKey(c.properties.KeyFile, c.properties.KeyPassword)
	if err != nil {
		panic(security.NewInternalError("cannot load private key from file", err))
	}
	rootURL, err := f.issuer.BuildUrl()
	if err != nil {
		panic(security.NewInternalError("cannot get issuer's base URL", err))
	}

	return Options{
		Key:         key,
		Cert:        cert[0],
		//usually this is the metadata url, but to keep consistent with existing implementation, we just use the context path
		EntityIdUrl: *rootURL,
		SsoUrl: *rootURL.ResolveReference(&url.URL{
			Path: fmt.Sprintf("%s%s", rootURL.Path, f.ssoLocation.Path),
			RawQuery: f.ssoLocation.RawQuery,
		}),
		serviceProviderManager: c.samlClientStore,
	}
}
