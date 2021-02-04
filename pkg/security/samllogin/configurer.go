package samllogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/dgrijalva/jwt-go"
	dsig "github.com/russellhaering/goxmldsig"
	"net/http"
	"net/url"
)

type SamlAuthConfigurer struct {
	properties   ServiceProviderProperties
	idpManager   idp.IdentityProviderManager
	serverProps  web.ServerProperties
	accountStore security.FederatedAccountStore
}

func (s *SamlAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	m := s.makeMiddleware(f, ws)

	ws.Route(matcher.RouteWithPattern(f.acsPath)).
		Route(matcher.RouteWithPattern(f.metadataPath)).
		Add(mapping.Get(f.metadataPath).
			HandlerFunc(m.MetadataHandlerFunc).
			Name("saml m metadata")).
		Add(mapping.Post(f.acsPath).
			HandlerFunc(m.ACSHandlerFunc).
			Name("saml assertion consumer m")).
		Add(middleware.NewBuilder("saml idp metadata refresh").
			Order(security.MWOrderSAMLMetadataRefresh).
			Use(m.RefreshMetadataHandler()))

	requestMatcher := matcher.RequestWithPattern(f.acsPath).Or(matcher.RequestWithPattern(f.metadataPath))
	access.Configure(ws).
	Request(requestMatcher).WithOrder(order.Highest).PermitAll()

	//authentication entry point
	errorhandling.Configure(ws).
		AuthenticationEntryPoint(request_cache.NewSaveRequestEntryPoint(m))
	return nil
}

func (s *SamlAuthConfigurer) effectiveSuccessHandler(f *Feature, ws security.WebSecurity) security.AuthenticationSuccessHandler {
	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler); ok {
		return security.NewAuthenticationSuccessHandler(globalHandler, f.successHandler)
	} else {
		return security.NewAuthenticationSuccessHandler(f.successHandler)
	}
}

func (s *SamlAuthConfigurer) getServiceProviderConfiguration(f *Feature) Options {
	cert, err := cryptoutils.LoadCert(s.properties.CertificateFile)
	if err != nil {
		panic(security.NewInternalError("cannot load certificate from file", err))
	}
	key, err := cryptoutils.LoadPrivateKey(s.properties.KeyFile, s.properties.KeyPassword)
	if err != nil {
		panic(security.NewInternalError("cannot load private key from file", err))
	}
	rootURL, err := url.Parse(s.properties.RootUrl)
	if err != nil {
		panic(security.NewInternalError("cannot parse security.auth.saml.root-url", err))
	}
	opts := Options{
		URL:            *rootURL,
		Key:            key,
		Certificate:    cert,
		ACSPath: 		fmt.Sprintf("%s%s", s.serverProps.ContextPath, f.acsPath),
		MetadataPath:   fmt.Sprintf("%s%s", s.serverProps.ContextPath, f.metadataPath),
		SLOPath: 		fmt.Sprintf("%s%s", s.serverProps.ContextPath, f.sloPath),
		SignRequest: true,
	}
	return opts
}

func (s *SamlAuthConfigurer) makeServiceProvider(opts Options) saml.ServiceProvider {
	metadataURL := opts.URL.ResolveReference(&url.URL{Path: opts.MetadataPath})
	acsURL := opts.URL.ResolveReference(&url.URL{Path: opts.ACSPath})
	sloURL := opts.URL.ResolveReference(&url.URL{Path: opts.SLOPath})

	var forceAuthn *bool
	if opts.ForceAuthn {
		forceAuthn = &opts.ForceAuthn
	}
	signatureMethod := dsig.RSASHA1SignatureMethod
	if !opts.SignRequest {
		signatureMethod = ""
	}

	sp := saml.ServiceProvider{
		Key:               opts.Key,
		Certificate:       opts.Certificate,
		Intermediates:     opts.Intermediates,
		MetadataURL:       *metadataURL,
		AcsURL:            *acsURL,
		SloURL:            *sloURL,
		ForceAuthn:        forceAuthn,
		SignatureMethod:   signatureMethod,
		AllowIDPInitiated: opts.AllowIDPInitiated,
	}
	return sp
}

func (s *SamlAuthConfigurer)  makeRequestTracker(opts Options) samlsp.RequestTracker {
	codec := samlsp.JWTTrackedRequestCodec{
		SigningMethod: jwt.SigningMethodRS256,
		Audience:      opts.URL.String(),
		Issuer:        opts.URL.String(),
		MaxAge:        saml.MaxIssueDelay,
		Key:           opts.Key,
	}

	//we want to set sameSite to none, which requires scheme to be https
	//otherwise we fallback to default mode, which on modern browsers is lax.
	//cross site functionality is limited in lax mode. the cookie will only be
	//sent cross site within 2 minutes of its creation.
	//so with none + https, we make sure production work as expected. and the fallback
	//provides limited support for development environment.
	secure := opts.URL.Scheme == "https"
	sameSite := http.SameSiteDefaultMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	tracker := CookieRequestTracker{
		NamePrefix:      "saml_",
		Codec:           codec,
		MaxAge:          saml.MaxIssueDelay,
		SameSite: 		 sameSite,
		Secure: 		 secure,
		Path:			 opts.ACSPath,
	}
	return tracker
}

func (s *SamlAuthConfigurer) makeMiddleware(f *Feature, ws security.WebSecurity) *ServiceProviderMiddleware {
	opts := s.getServiceProviderConfiguration(f)
	sp := s.makeServiceProvider(opts)
	tracker := s.makeRequestTracker(opts)
	if f.successHandler == nil {
		f.successHandler = NewTrackedRequestSuccessHandler(tracker)
	}

	authenticator := &Authenticator{
		accountStore: s.accountStore,
		idpManager: s.idpManager,
	}

	clientManager := NewCacheableIdpClientManager(sp)

	return NewMiddleware(sp, tracker, s.idpManager, clientManager, s.effectiveSuccessHandler(f, ws), authenticator, f.errorPath)
}

func newSamlAuthConfigurer(properties ServiceProviderProperties, serverProps web.ServerProperties, idpManager idp.IdentityProviderManager,
	accountStore security.FederatedAccountStore) *SamlAuthConfigurer {
	return &SamlAuthConfigurer{
		properties: properties,
		idpManager: idpManager,
		serverProps: serverProps,
		accountStore: accountStore,
	}
}