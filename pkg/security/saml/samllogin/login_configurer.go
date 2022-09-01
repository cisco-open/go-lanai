package samllogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
)

type SamlAuthConfigurer struct {
	*samlConfigurer
	accountStore   security.FederatedAccountStore
}

func (c *SamlAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	m := c.makeMiddleware(f, ws)

	ws.Route(matcher.RouteWithPattern(f.acsPath)).
		Route(matcher.RouteWithPattern(f.metadataPath)).
		Add(mapping.Get(f.metadataPath).
			HandlerFunc(m.MetadataHandlerFunc()).
			//metadata is an endpoint that is available without conditions, therefore call Build() to not inherit the ws condition
			Name("saml metadata").Build()).
		Add(mapping.Post(f.acsPath).
			HandlerFunc(m.ACSHandlerFunc()).
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

func (c *SamlAuthConfigurer) makeMiddleware(f *Feature, ws security.WebSecurity) *SPLoginMiddleware {
	opts := c.getServiceProviderConfiguration(f)
	sp := c.sharedServiceProvider(opts)
	clientManager := c.sharedClientManager(opts)
	tracker := c.sharedRequestTracker(opts)
	if f.successHandler == nil {
		f.successHandler = NewTrackedRequestSuccessHandler(tracker)
	}

	authenticator := &Authenticator{
		accountStore: c.accountStore,
		idpManager:   c.samlIdpManager,
	}


	return NewLoginMiddleware(sp, tracker, c.idpManager, clientManager, c.effectiveSuccessHandler(f, ws), authenticator, f.errorPath)
}

func newSamlAuthConfigurer(shared *samlConfigurer, accountStore security.FederatedAccountStore) *SamlAuthConfigurer {
	return &SamlAuthConfigurer{
		samlConfigurer: shared,
		accountStore:   accountStore,
	}
}