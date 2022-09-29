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
	handlers := make([]security.AuthenticationSuccessHandler, len(f.successHandlers), len(f.successHandlers) + 2)
	copy(handlers, f.successHandlers)

	if len(handlers) == 0 {
		handlers = append(handlers, redirect.NewRedirectWithURL(f.successUrl))
	}

	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler); ok {
		handlers = append([]security.AuthenticationSuccessHandler{globalHandler}, handlers...) // global BEFORE logout success handlers
	}
	order.SortStable(handlers, order.OrderedFirstCompare)
	return security.NewAuthenticationSuccessHandler(handlers...)
}

func (c *LogoutConfigurer) effectiveErrorHandler(f *LogoutFeature, ws security.WebSecurity) security.AuthenticationErrorHandler {
	handlers := make([]security.AuthenticationErrorHandler, len(f.errorHandlers), len(f.errorHandlers) + 2)
	copy(handlers, f.errorHandlers)

	if len(handlers) == 0 {
		handlers = append(handlers, redirect.NewRedirectWithURL(f.errorUrl))
	}

	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthErrorHandler).(security.AuthenticationErrorHandler); ok {
		handlers = append(handlers, globalHandler) // global AFTER logout error handlers
	}
	return security.NewAuthenticationErrorHandler(handlers...)
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


