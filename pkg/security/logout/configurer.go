package logout

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
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

func (c *LogoutConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Verify
	if err := c.validate(feature.(*LogoutFeature), ws); err != nil {
		return err
	}
	f := feature.(*LogoutFeature)

	supportedMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}
	// let ws know to intercept additional url
	route := matcher.RouteWithPattern(f.logoutUrl, supportedMethods...)
	ws.Route(route)

	// configure middlewares
	// Note: since this MW handles a new path, we add middleware as-is instead of a security.MiddlewareTemplate
	order.SortStable(f.logoutHandlers, order.OrderedFirstCompare)
	logout := NewLogoutMiddleware(
		c.effectiveSuccessHandler(f, ws),
		c.effectiveErrorHandler(f, ws),
		c.effectiveEntryPoints(f),
		f.logoutHandlers...)
	mw := middleware.NewBuilder("logout").
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

func (c *LogoutConfigurer) validate(f *LogoutFeature, _ security.WebSecurity) error {
	if f.logoutUrl == "" {
		return fmt.Errorf("logoutUrl is missing for logout")
	}

	if f.successUrl == "" && len(f.successHandlers) == 0 {
		return fmt.Errorf("successUrl and successHandler are both missing for logout")
	}

	return nil
}

func (c *LogoutConfigurer) effectiveSuccessHandler(f *LogoutFeature, ws security.WebSecurity) security.AuthenticationSuccessHandler {

	if len(f.successHandlers) == 0 {
		f.successHandlers = []security.AuthenticationSuccessHandler{redirect.NewRedirectWithURL(f.successUrl)}
	}

	order.SortStable(f.successHandlers, order.OrderedFirstCompare)
	sh := security.NewAuthenticationSuccessHandler(f.successHandlers...)
	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler); ok {
		return security.NewAuthenticationSuccessHandler(globalHandler, sh)
	} else {
		return sh
	}
}

func (c *LogoutConfigurer) effectiveErrorHandler(f *LogoutFeature, ws security.WebSecurity) security.AuthenticationErrorHandler {

	if len(f.errorHandlers) == 0 {
		f.errorHandlers = []security.AuthenticationErrorHandler{redirect.NewRedirectWithURL(f.errorUrl)}
	}

	order.SortStable(f.errorHandlers, order.OrderedFirstCompare)
	eh := security.NewAuthenticationErrorHandler(f.errorHandlers...)
	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthErrorHandler).(security.AuthenticationErrorHandler); ok {
		return security.NewAuthenticationErrorHandler(globalHandler, eh)
	} else {
		return eh
	}
}

func (c *LogoutConfigurer) effectiveEntryPoints(f *LogoutFeature) security.AuthenticationEntryPoint {
	if len(f.entryPoints) == 0 {
		return nil
	}

	order.SortStable(f.entryPoints, order.OrderedFirstCompare)
	return multiEntryPoints(f.entryPoints)
}

type multiEntryPoints []security.AuthenticationEntryPoint

func (ep multiEntryPoints) Commence(ctx context.Context, request *http.Request, writer http.ResponseWriter, err error) {
	for _, entryPoint := range ep {
		entryPoint.Commence(ctx, request, writer, err)
	}
}


