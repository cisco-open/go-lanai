package saml_auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
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

func newSamlAuthorizeEndpointConfigurer(properties samlctx.SamlProperties, serverProperties web.ServerProperties,
	samlClientStore SamlClientStore,
	accountStore security.AccountStore,
	attributeGenerator AttributeGenerator) *SamlAuthorizeEndpointConfigurer {

	return &SamlAuthorizeEndpointConfigurer{
		properties:         properties,
		serverProperties:   serverProperties,
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
	key, err := cryptoutils.LoadPrivateKey(c.properties.KeyFile, c.properties.KeyPassword)
	if err != nil {
		panic(security.NewInternalError("cannot load private key from file", err))
	}
	rootURL, err := url.Parse(c.properties.RootUrl)
	if err != nil {
		panic(security.NewInternalError("cannot parse security.auth.saml.root-url", err))
	}

	return Options{
		Key:         key,
		Cert:        cert,
		//usually this is the metadata url, but to keep consistent with existing implementation, we just use the context path
		EntityIdUrl: *rootURL.ResolveReference(&url.URL{Path: c.serverProperties.ContextPath}),
		SsoUrl: *rootURL.ResolveReference(&url.URL{
			Path: fmt.Sprintf("%s%s", c.serverProperties.ContextPath, f.ssoLocation.Path),
			RawQuery: f.ssoLocation.RawQuery,
		}),
		serviceProviderManager: c.samlClientStore,
	}
}
