package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
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
	// Verify
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

	if err := flc.configureMfaPage(f, ws); err != nil {
		return err
	}

	if err := flc.configureLoginProcessing(f, ws); err != nil {
		return err
	}

	if err := flc.configureMfaProcessing(f, ws); err != nil {
		return err
	}

	//TODO

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
		f.loginErrorUrl = f.loginUrl + "?error=true"
	}

	if f.mfaEnabled && f.mfaUrl == "" {
		return fmt.Errorf("mfaUrl is missing for MFA")
	}

	if f.mfaEnabled && f.mfaSuccessUrl == "" && f.successHandler == nil {
		f.mfaSuccessUrl = f.loginSuccessUrl
	}

	if f.mfaEnabled &&  f.mfaVerifyUrl == "" {
		f.mfaVerifyUrl = f.mfaUrl
	}

	if f.mfaErrorUrl == "" && f.failureHandler == nil {
		f.mfaErrorUrl = f.mfaUrl + "?error=true"
	}

	if f.formProcessCondition == nil {
		if wsReader, ok := ws.(security.WebSecurityReader); ok {
			f.formProcessCondition = wsReader.GetCondition()
		} else {
			return fmt.Errorf("formProcessCondition is not specified and unable to read condition from WebSecurity")
		}
	}

	return nil
}

func (flc *FormLoginConfigurer) configureErrorHandling(f *FormLoginFeature, ws security.WebSecurity) error {
	errorRedirect := redirect.NewRedirectWithURL(f.loginErrorUrl)
	mfaErrorRedirect := redirect.NewRedirectWithURL(f.mfaErrorUrl)

	if f.failureHandler == nil {
		f.failureHandler = errorRedirect
	}

	var entryPoint security.AuthenticationEntryPoint = redirect.NewRedirectWithURL(f.loginUrl)
	if f.mfaEnabled {
		if _, ok := f.failureHandler.(*MfaAwareAuthenticationErrorHandler); !ok {
			f.failureHandler = &MfaAwareAuthenticationErrorHandler{
				delegate: f.failureHandler,
				mfaPendingDelegate: mfaErrorRedirect,
			}
		}

		entryPoint = &MfaAwareAuthenticationEntryPoint {
			delegate: entryPoint,
			mfaPendingDelegate: redirect.NewRedirectWithURL(f.mfaUrl),
		}
	}

	// override entry point and error handler
	errorhandling.Configure(ws).
		AuthenticationEntryPoint(session.NewSaveRequestEntryPoint(entryPoint)).
		AuthenticationErrorHandler(f.failureHandler)

	// adding CSRF protection err handler, while keeping default
	csrf.Configure(ws).CsrfDeniedHandler(errorRedirect)

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

func (flc *FormLoginConfigurer) configureMfaPage(f *FormLoginFeature, ws security.WebSecurity) error {
	// let ws know to intercept additional url
	routeMatcher := matcher.RouteWithPattern(f.mfaUrl, http.MethodGet)
	requestMatcher := matcher.RequestWithPattern(f.mfaUrl, http.MethodGet)
	ws.Route(routeMatcher)

	// configure access
	access.Configure(ws).
		Request(requestMatcher).WithOrder(order.Highest).
		HasPermissions(passwd.SpecialPermissionMFAPending, passwd.SpecialPermissionOtpId)

	return nil
}

func (flc *FormLoginConfigurer) configureLoginProcessing(f *FormLoginFeature, ws security.WebSecurity) error {

	// let ws know to intercept additional url
	route := matcher.RouteWithPattern(f.loginProcessUrl, http.MethodPost)
	ws.Route(route)

	// configure middlewares
	// Note: since this MW handles a new path, we add middleware as-is instead of a security.MiddlewareTemplate

	login := NewFormAuthenticationMiddleware(func(opts *FormAuthMWOptions) {
		opts.Authenticator = ws.Authenticator()
		opts.SuccessHandler = flc.effectiveSuccessHandler(f, ws)
		opts.UsernameParam =  f.usernameParam
		opts.PasswordParam = f.passwordParam
	})
	mw := middleware.NewBuilder("form login").
		ApplyTo(route).
		WithCondition(security.WebConditionFunc(f.formProcessCondition)).
		Order(security.MWOrderFormAuth).
		Use(login.LoginProcessHandlerFunc()).
		Build()

	ws.Add(mw)

	// configure additional endpoint mappings to trigger middleware
	ws.Add(web.NewGenericMapping("login process dummy", f.loginProcessUrl, http.MethodPost, login.EndpointHandlerFunc() ))

	return nil
}

func (flc *FormLoginConfigurer) configureMfaProcessing(f *FormLoginFeature, ws security.WebSecurity) error {

	// let ws know to intercept additional url
	routeVerify := matcher.RouteWithPattern(f.mfaVerifyUrl, http.MethodPost)
	routeRefresh :=	matcher.RouteWithPattern(f.mfaRefreshUrl, http.MethodPost)
	requestMatcher := matcher.RequestWithPattern(f.mfaRefreshUrl, http.MethodPost).
		Or(matcher.RequestWithPattern(f.mfaRefreshUrl, http.MethodPost))
	ws.Route(routeVerify).Route(routeRefresh)

	// configure middlewares
	// Note: since this MW handles a new path, we add middleware as-is instead of a security.MiddlewareTemplate
	login := NewMfaAuthenticationMiddleware(func(opts *MfaMWOptions) {
		opts.Authenticator = ws.Authenticator()
		opts.SuccessHandler = flc.effectiveSuccessHandler(f, ws)
		opts.OtpParam =  f.otpParam
	})

	verifyMW := middleware.NewBuilder("otp verify").
		ApplyTo(routeVerify).
		WithCondition(security.WebConditionFunc(f.formProcessCondition)).
		Order(security.MWOrderFormAuth).
		Use(login.OtpVerifyHandlerFunc()).
		Build()

	refreshMW := middleware.NewBuilder("otp refresh").
		ApplyTo(routeRefresh).
		WithCondition(security.WebConditionFunc(f.formProcessCondition)).
		Order(security.MWOrderFormAuth).
		Use(login.OtpRefreshHandlerFunc()).
		Build()

	ws.Add(verifyMW, refreshMW)

	// configure additional endpoint mappings to trigger middleware
	ws.Add(web.NewGenericMapping("otp verify dummy", f.mfaVerifyUrl, http.MethodPost, login.EndpointHandlerFunc()) )
	ws.Add(web.NewGenericMapping("otp refresh dummy", f.mfaRefreshUrl, http.MethodPost, login.EndpointHandlerFunc()) )

	// configure access
	access.Configure(ws).
		Request(requestMatcher).WithOrder(order.Highest).
		HasPermissions(passwd.SpecialPermissionMFAPending, passwd.SpecialPermissionOtpId)

	return nil
}

func (flc *FormLoginConfigurer) configureCSRF(f *FormLoginFeature, ws security.WebSecurity) error {
	csrfMatcher := matcher.RequestWithPattern(f.loginProcessUrl, http.MethodPost).
		Or(matcher.RequestWithPattern(f.mfaVerifyUrl, http.MethodPost)).
		Or(matcher.RequestWithPattern(f.mfaRefreshUrl, http.MethodPost))
	csrf.Configure(ws).AddCsrfProtectionMatcher(csrfMatcher)
	return nil
}

func (flc *FormLoginConfigurer) effectiveSuccessHandler(f *FormLoginFeature, ws security.WebSecurity) security.AuthenticationSuccessHandler {

	if f.successHandler == nil {
		f.successHandler = session.NewSavedRequestAuthenticationSuccessHandler(redirect.NewRedirectWithURL(f.loginSuccessUrl))
	}

	if _, ok := f.successHandler.(*MfaAwareSuccessHandler); f.mfaEnabled && !ok {
		f.successHandler = &MfaAwareSuccessHandler{
			delegate: f.successHandler,
			mfaPendingDelegate: redirect.NewRedirectWithURL(f.mfaUrl),
		}
	}


	if globalHandler, ok := ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(security.AuthenticationSuccessHandler); ok {
		return security.NewAuthenticationSuccessHandler(globalHandler, f.successHandler)
	} else {
		return f.successHandler
	}
}