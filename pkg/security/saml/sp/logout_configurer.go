package sp

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
		AddEntryPoint(ep)

	// Add some additional endpoints.
	// Note: those endpoints are available regardless what auth method is used, so no condition is applied
	ws.Route(matcher.RouteWithPattern(f.sloPath)).
		Add(mapping.Get(f.sloPath).
			HandlerFunc(m.LogoutHandlerFunc()).
			Name("saml slo as sp - get"),
		).
		Add(mapping.Post(f.sloPath).
			HandlerFunc(m.LogoutHandlerFunc()).
			Name("saml slo as sp - post"),
		).
		Add(middleware.NewBuilder("saml idp metadata refresh").
			Order(security.MWOrderSAMLMetadataRefresh).
			Use(m.RefreshMetadataHandler()),
		)

	csrf.Configure(ws).
		IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(f.sloPath))
	return nil
}

func (c *SamlLogoutConfigurer) makeLogoutHandler(_ *Feature, _ security.WebSecurity) *SingleLogoutHandler {
	return NewSingleLogoutHandler()
}

func (c *SamlLogoutConfigurer) makeMiddleware(f *Feature, ws security.WebSecurity) *SPLogoutMiddleware {
	opts := c.getServiceProviderConfiguration(f)
	sp := c.sharedServiceProvider(opts)
	clientManager := c.sharedClientManager(opts)
	if f.successHandler == nil {
		f.successHandler = request_cache.NewSavedRequestAuthenticationSuccessHandler(
			redirect.NewRedirectWithURL("/"),
			func(from, to security.Authentication) bool {
				return true
			},
		)
	}

	return NewLogoutMiddleware(sp, c.idpManager, clientManager, c.effectiveSuccessHandler(f, ws))
}

func newSamlLogoutConfigurer(shared *samlConfigurer) *SamlLogoutConfigurer {
	return &SamlLogoutConfigurer{
		samlConfigurer: shared,
	}
}
