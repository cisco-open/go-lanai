package logout

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/mapping"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
)

var (
	FeatureId = security.FeatureId("Logout", security.FeatureOrderLogout)
)

//goland:noinspection GoNameStartsWithPackageName
type LogoutConfigurer struct {

}

func newLogoutConfigurer() *LogoutConfigurer {
	return &LogoutConfigurer{
	}
}

func (flc *LogoutConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Verify
	if err := flc.validate(feature.(*LogoutFeature), ws); err != nil {
		return err
	}
	f := feature.(*LogoutFeature)

	supportedMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
	}
	// let ws know to intercept additional url
	route := matcher.RouteWithPattern(f.logoutUrl, supportedMethods...)
	ws.Route(route)

	// configure middlewares
	// Note: since this MW handles a new path, we add middleware as-is instead of a security.MiddlewareTemplate
	logout := NewLogoutMiddleware(
		flc.effectiveSuccessHandler(f, ws),
		flc.effectiveErrorHandler(f, ws),
		f.logoutHandlers...)
	mw := middleware.NewBuilder("form logout").
		ApplyTo(route).
		Order(security.MWOrderFormLogout).
		Use(logout.LogoutHandlerFunc())

	ws.Add(mw)

	// configure additional endpoint mappings to trigger middleware
	for _,method := range supportedMethods {
		endpoint := mapping.New("logout dummy " + method).
			Method(method).Path(f.logoutUrl).
			HandlerFunc(security.NoopHandlerFunc())
		ws.Add(endpoint)
	}
	return nil
}

func (flc *LogoutConfigurer) validate(f *LogoutFeature, ws security.WebSecurity) error {
	if f.logoutUrl == "" {
		return fmt.Errorf("logoutUrl is missing for logout")
	}

	if f.successUrl == "" && f.successHandler == nil {
		return fmt.Errorf("successUrl and successHandler are both missing for logout")
	}

	return nil
}

func (flc *LogoutConfigurer) effectiveSuccessHandler(f *LogoutFeature, ws security.WebSecurity) security.AuthenticationSuccessHandler {

	if f.successHandler == nil {
		f.successHandler = redirect.NewRedirectWithURL(f.successUrl)
	}

	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler); ok {
		return security.NewAuthenticationSuccessHandler(globalHandler, f.successHandler)
	} else {
		return f.successHandler
	}
}

func (flc *LogoutConfigurer) effectiveErrorHandler(f *LogoutFeature, ws security.WebSecurity) security.AuthenticationErrorHandler {

	if f.errorHandler == nil {
		f.errorHandler = redirect.NewRedirectWithURL(f.errorUrl)
	}

	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthErrorHandler).(security.AuthenticationErrorHandler); ok {
		return security.NewAuthenticationErrorHandler(globalHandler, f.errorHandler)
	} else {
		return f.errorHandler
	}
}