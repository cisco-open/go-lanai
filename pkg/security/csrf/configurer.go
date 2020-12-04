package csrf

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

/**
	CSRF feature uses the synchronizer token pattern to prevent cross site request forgery
	https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#synchronizer-token-pattern
 */

var FeatureId = security.FeatureId("csrf", security.FeatureOrderCsrf)

type Feature struct {
	RequireCsrfProtectionMatcher web.RequestMatcher
	csrfDeniedHandler security.AccessDeniedHandler
}

func Configure(ws security.WebSecurity) *Feature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return  fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure CSRF: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

func New() *Feature {
	return &Feature{}
}

func (f *Feature) CsrfProtectionMatcher(m web.RequestMatcher) *Feature {
	f.RequireCsrfProtectionMatcher = m
	return f
}

func (f *Feature) CsrfDeniedHandler(csrfDeniedHandler security.AccessDeniedHandler) *Feature {
	f.csrfDeniedHandler = csrfDeniedHandler
	return f
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

type Configurer struct {
}

func newCsrfConfigurer() *Configurer{
	return &Configurer{}
}

func (sc *Configurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	// configure additional access denied handler if provided
	if f.csrfDeniedHandler != nil {
		handler := &CsrfDeniedHandler{delegate: f.csrfDeniedHandler}
		ws.Shared(security.WSSharedKeyCompositeAccessDeniedHandler).(*security.CompositeAccessDeniedHandler).
			Add(handler)
	}

	// configure middleware
	tokenStore := newSessionBackedStore()
	manager := newManager(tokenStore, f.RequireCsrfProtectionMatcher)
	csrfHandler := middleware.NewBuilder("csrfMiddleware").
		Order(security.MWOrderCsrfHandling).
		Use(manager.CsrfHandlerFunc())

	ws.Add(csrfHandler)
	return nil
}

