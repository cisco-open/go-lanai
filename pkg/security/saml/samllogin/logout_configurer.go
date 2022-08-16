package samllogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
)

type SamlLogoutConfigurer struct {
	*samlConfigurer
}

func (c *SamlLogoutConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	m := c.makeMiddleware(f, ws)
	lh := c.makeLogoutHandler(f, ws)
	ep := request_cache.NewSaveRequestEntryPoint(m)

	// configure on top of existing logout feature
	logout.Configure(ws).
		AddLogoutHandler(lh).
		EntryPoint(ep)

	// Add some additional endpoints.
	// Note: those endpoints are available regardless what auth method is used, so no condition is applied
	// TODO make it configurable
	ws.Route(matcher.RouteWithPattern(f.sloPath)).
		Route(matcher.RouteWithPattern("/v2/logout/saml/slo/dummy")).
		Add(mapping.Get(f.sloPath).
			HandlerFunc(m.LogoutRequestHandlerFunc()).
			Name("saml slo as sp").Build(),
		).
		Add(mapping.Post(f.sloPath).
			HandlerFunc(m.LogoutResponseHandlerFunc()).
			Name("saml slo callback as sp"),
		).
		Add(mapping.Get("/v2/logout/saml/slo/dummy").
			HandlerFunc(m.DummySLOHandlerFunc()).
			Name("dummy saml slo as sp").Build(),
		).
		Add(middleware.NewBuilder("saml idp metadata refresh").
			Order(security.MWOrderSAMLMetadataRefresh).
			Use(m.RefreshMetadataHandler()),
		)

	csrf.Configure(ws).
		IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(f.sloPath))

	// TODO In case SLO endpoints are invoked when there is no active authenticated session, security would entry point
	// 		to handle this error. We need to configure it properly
	//errorhandling.Configure(ws).
	//	AuthenticationEntryPoint(request_cache.NewSaveRequestEntryPoint(m))
	return nil
}

func (c *SamlLogoutConfigurer) makeLogoutHandler(f *Feature, ws security.WebSecurity) *SingleLogoutHandler {
	// TODO review this part
	return NewSingleLogoutHandler()
}

func (c *SamlLogoutConfigurer) makeMiddleware(f *Feature, ws security.WebSecurity) *SPLogoutMiddleware {
	// TODO revise this part
	opts := c.getServiceProviderConfiguration(f)
	sp := c.sharedServiceProvider(opts)
	clientManager := c.sharedClientManager(opts)
	tracker := c.sharedRequestTracker(opts)
	if f.successHandler == nil {
		f.successHandler = request_cache.NewSavedRequestAuthenticationSuccessHandler(
			redirect.NewRedirectWithURL("/"),
			func(from, to security.Authentication) bool {
				return true
			},
		)
	}

	return NewLogoutMiddleware(sp, c.idpManager, clientManager, tracker, c.effectiveSuccessHandler(f, ws), f.errorPath)
}

func newSamlLogoutConfigurer(shared *samlConfigurer) *SamlLogoutConfigurer {
	return &SamlLogoutConfigurer{
		samlConfigurer: shared,
	}
}
