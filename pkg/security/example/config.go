package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/formlogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"fmt"
	"time"
)

type TestSecurityConfigurer struct {
	accountStore security.AccountStore
}

func (c *TestSecurityConfigurer) Configure(ws security.WebSecurity) {

	// DSL style example
	// for REST API
	ws.Route(matcher.RouteWithPattern("/api/**")).
		Condition(matcher.RequestWithHost("localhost:8080")).
		//With(session.New()).
		With(passwd.New().
			AccountStore(c.accountStore).
			PasswordEncoder(passwd.NewNoopPasswordEncoder()),
		).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(basicauth.New()).
		With(errorhandling.New())
}


type AnotherSecurityConfigurer struct {
}

func (c *AnotherSecurityConfigurer) Configure(ws security.WebSecurity) {

	// For Page
	handler := redirect.NewRedirectWithRelativePath("/error")
	condition := matcher.RequestWithHost("localhost:8080")

	ws.Route(matcher.RouteWithPattern("/page/**")).
		Condition(condition).
		With(session.New()).
		With(passwd.New().
			MFA(true).
			OtpTTL(5 * time.Minute).
			MFAEventListeners(debugPrintOTP),
		).
		With(access.New().
			Request(
				matcher.RequestWithPattern("/page/public").
					Or(matcher.RequestWithPattern("/page/public/**")),
			).PermitAll().
			Request(matcher.AnyRequest()).HasPermissions("welcomed"),
		).
		With(formlogin.New().
			FormProcessCondition(condition).
			EnableMFA(),
		).
		With(logout.New().
			SuccessUrl("/login"),
		).
		With(errorhandling.New().
			AuthenticationEntryPoint(handler).
			AccessDeniedHandler(handler),
		).
		With(csrf.New().IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern("/page/process")))
}

type ErrorPageSecurityConfigurer struct {
}

func (c *ErrorPageSecurityConfigurer) Configure(ws security.WebSecurity) {

	ws.Route(matcher.RouteWithPattern("/error")).
		With(session.New()).
		With(access.New().
			Request(matcher.AnyRequest()).PermitAll(),
		)
}

func debugPrintOTP(event passwd.MFAEvent, otp passwd.OTP, principal interface{}) {
	switch event {
	case passwd.MFAEventOtpCreate, passwd.MFAEventOtpRefresh:
		fmt.Printf("[DEBUG] OTP: %s \n", otp.Passcode())
	}
}
