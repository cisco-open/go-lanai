package logout

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
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
	// Validate
	if err := flc.validate(feature.(*LogoutFeature), ws); err != nil {
		return err
	}
	f := feature.(*LogoutFeature)

	if f.successHandler == nil {
		f.successHandler = redirect.NewRedirectWithURL(f.successUrl)
	}
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
	logout := NewLogoutMiddleware(f.successHandler, f.logoutHandlers...)
	mw := middleware.NewBuilder("form logout").
		ApplyTo(route).
		WithCondition(security.WebConditionFunc(f.condition)).
		Order(security.MWOrderFormLogout).
		Use(logout.LogoutHandlerFunc()).
		Build()

	ws.Add(mw)

	// configure additional endpoint mappings to trigger middleware
	for _,method := range supportedMethods {
		endpoint := web.NewGenericMapping("logout dummy " + method, f.logoutUrl, method, logout.EndpointHandlerFunc())
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

	if f.condition == nil {
		if wsReader, ok := ws.(security.WebSecurityReader); ok {
			f.condition = wsReader.GetCondition()
		} else {
			return fmt.Errorf("condition is not specified and unable to read condition from WebSecurity")
		}
	}

	return nil
}