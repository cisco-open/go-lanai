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

	if err := flc.configureLoginPage(f, ws); err != nil {
		return err
	}

	if err := flc.configureLoginProcessing(f, ws); err != nil {
		return err
	}

	return nil
}

func (flc *FormLoginConfigurer) validate(f *FormLoginFeature, ws security.WebSecurity) error {
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

	if f.loginProcessCondition == nil {
		if wsReader, ok := ws.(security.WebSecurityReader); ok {
			f.loginProcessCondition = wsReader.GetCondition()
		} else {
			return fmt.Errorf("loginProcessCondition is not specified and unable to read condition from WebSecurity")
		}
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

func (flc *FormLoginConfigurer) configureLoginPage(f *FormLoginFeature, ws security.WebSecurity) error {
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
	route := matcher.RouteWithPattern(f.loginProcessUrl, http.MethodPost)
	ws.Route(route)

	// configure middlewares
	// Note: since this MW handles a new path, we add middleware as-is instead of a security.MiddlewareTemplate

	login := NewFormAuthenticationMiddleware(FormAuthOptions{
		Authenticator:  ws.Authenticator(),
		SuccessHandler: f.successHandler,
		UsernameParam:  f.usernameParam,
		PasswordParam:  f.passwordParam,
	})
	mw := middleware.NewBuilder("form login").
		ApplyTo(route).
		WithCondition(security.WebConditionFunc(f.loginProcessCondition)).
		Order(security.MWOrderFormAuth).
		Use(login.LoginProcessHandlerFunc()).
		Build()

	ws.Add(mw)

	// configure additional endpoint mappings to trigger middleware
	ws.Add(web.NewGenericMapping("login process dummy", f.loginProcessUrl, http.MethodPost, login.EmptyHandlerFunc() ))
	return nil
}