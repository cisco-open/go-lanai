package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
)

const (
	wsSharedKeyLoginProcessEndpoint = "login process endpoint"
	wsSharedKeyLogoutEndpoint = "logout endpoint"
)

var (
	FeatureId = security.FeatureId("FormLogin", security.FeatureOrderFormLogin)
)

//goland:noinspection GoNameStartsWithPackageName
type FormLoginConfigurer struct {

}

func newFormLoginConfigurer() *FormLoginConfigurer {
	return &FormLoginConfigurer{
	}
}

func (flc *FormLoginConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Validate
	if err := flc.validate(feature.(*FormLoginFeature), ws); err != nil {
		return err
	}
	f := feature.(*FormLoginFeature)


	if err := flc.configureErrorHandling(f, ws); err != nil {
		return err
	}

	if err := flc.configurePageAccessControl(f, ws); err != nil {
		return err
	}

	if err := flc.configureLoginProcessing(f, ws); err != nil {
		return err
	}

	if err := flc.configureLogout(f, ws); err != nil {
		return err
	}

	return nil
}

func (flc *FormLoginConfigurer) validate(f *FormLoginFeature, _ security.WebSecurity) error {
	if f.loginUrl == "" {
		return fmt.Errorf("loginUrl is missing for form login")
	}

	if f.loginSuccessUrl == "" && f.successHandler == nil {
		return fmt.Errorf("loginSuccessUrl and successHanlder are missing for form login")
	}

	if f.loginProcessUrl == "" {
		f.loginProcessUrl = f.loginUrl
	}

	if f.loginErrorUrl == "" && f.failureHandler == nil {
		f.loginErrorUrl = f.loginUrl + "?error"
	}

	return nil
}

func (flc *FormLoginConfigurer) configureErrorHandling(f *FormLoginFeature, ws security.WebSecurity) error {
	if f.failureHandler == nil {
		f.failureHandler = redirect.NewRedirectWithURL(f.loginErrorUrl)
	}

	errorhandling.Configure(ws).
		AuthenticationEntryPoint(redirect.NewRedirectWithURL(f.loginUrl)).
		AuthenticationErrorHandler(f.failureHandler)

	return nil
}

func (flc *FormLoginConfigurer) configurePageAccessControl(f *FormLoginFeature, ws security.WebSecurity) error {

	// let ws know to intercept additional url
	routeMatcher := matcher.RouteWithPattern(f.loginUrl, http.MethodGet)
	requestMatcher := matcher.RequestWithPattern(f.loginUrl, http.MethodGet)
	ws.Route(routeMatcher)

	// configure access
	access.Configure(ws).
		Request(requestMatcher).WithOrder(order.Highest).PermitAll()

	return nil
}

func (flc *FormLoginConfigurer) configureLoginProcessing(f *FormLoginFeature, ws security.WebSecurity) error {
	if f.successHandler == nil {
		f.successHandler = redirect.NewRedirectWithURL(f.loginSuccessUrl)
	}

	// let ws know to intercept additional url
	routeMatcher := matcher.RouteWithPattern(f.loginProcessUrl, http.MethodPost)
	requestMatcher := matcher.RequestWithPattern(f.loginProcessUrl, http.MethodPost)
	ws.Route(routeMatcher)

	// configure middlewares
	login := NewFormAuthenticationMiddleware(FormAuthOptions{
		Authenticator:  ws.Authenticator(),
		SuccessHandler: f.successHandler,
		UsernameParam:  f.usernameParam,
		PasswordParam:  f.passwordParam,
		RequestMatcher: requestMatcher,
	})
	auth := middleware.NewBuilder("form login").
		Order(security.MWOrderFormAuth).
		Use(login.LoginProcessHandlerFunc())

	ws.Add(auth)

	// configure additional endpoint mappings to trigger middleware
	err := ws.AddShared(wsSharedKeyLoginProcessEndpoint,
		web.NewGenericMapping(wsSharedKeyLoginProcessEndpoint, f.loginProcessUrl, http.MethodPost, login.EmptyHandlerFunc()))
	return err
}

func (flc *FormLoginConfigurer) configureLogout(f *FormLoginFeature, ws security.WebSecurity) error {
	// let ws know to intercept additional url
	routeMatcher := matcher.RouteWithPattern(f.logoutUrl, http.MethodGet, http.MethodPost)
	requestMatcher := matcher.RequestWithPattern(f.logoutUrl, http.MethodGet, http.MethodPost)
	ws.Route(routeMatcher)
	// TODO need a endpoint to trigger this mapping

	// configure middlewares
	login := NewFormAuthenticationMiddleware(FormAuthOptions{
		RequestMatcher: requestMatcher,
	})
	auth := middleware.NewBuilder("form logout").
		Order(security.MWOrderFormLogout).
		Use(login.LoginProcessHandlerFunc())

	ws.Add(auth)

	// configure additional endpoint mappings to trigger middleware
	err := ws.AddShared(wsSharedKeyLogoutEndpoint + " Get",
		web.NewGenericMapping(wsSharedKeyLogoutEndpoint, f.logoutUrl, http.MethodGet, login.EmptyHandlerFunc()))

	err = ws.AddShared(wsSharedKeyLogoutEndpoint + " Post",
		web.NewGenericMapping(wsSharedKeyLogoutEndpoint, f.logoutUrl, http.MethodPost, login.EmptyHandlerFunc()))
	return err
}
