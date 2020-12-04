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

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

type Configurer struct {
}

func newCsrfConfigurer() *Configurer{
	return &Configurer{}
}

func (sc *Configurer) Apply(f security.Feature, ws security.WebSecurity) error {

	tokenStore := newSessionBackedStore()
	manager := newManager(tokenStore, f.(*Feature).RequireCsrfProtectionMatcher)
	csrfHandler := middleware.NewBuilder("csrfMiddleware").
		Order(security.MWOrderCsrfHandling).
		Use(manager.CsrfHandlerFunc())

	ws.Add(csrfHandler)
	return nil
}
