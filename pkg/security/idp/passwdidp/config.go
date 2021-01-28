package passwdidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/formlogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"fmt"
	"time"
)

type PasswordIdpSecurityConfigurer struct {

}

func NewPasswordIdpSecurityConfigurer() *PasswordIdpSecurityConfigurer {
	return &PasswordIdpSecurityConfigurer{

	}
}

func (c *PasswordIdpSecurityConfigurer) ConfigureAuthorizeEndpoint(path string, ws security.WebSecurity) {
	// TODO
	// For Authorize endpoint
	handler := redirect.NewRedirectWithRelativePath("/error")
	condition := matcher.RequestWithHost("internal.vms.com:8080")

	ws.Condition(condition).
		With(session.New()).
		With(passwd.New().
			MFA(true).
			OtpTTL(5 * time.Minute).
			MFAEventListeners(debugPrintOTP),
		).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(formlogin.New().
			EnableMFA(),
		).
		//With(logout.New().
		//	SuccessUrl("/page/hello"),
		//).
		With(errorhandling.New().
			AuthenticationEntryPoint(handler).
			AccessDeniedHandler(handler),
		).
		With(csrf.New().
			IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(path)),
		).
		With(request_cache.New())
}

func debugPrintOTP(event passwd.MFAEvent, otp passwd.OTP, principal interface{}) {
	switch event {
	case passwd.MFAEventOtpCreate, passwd.MFAEventOtpRefresh:
		fmt.Printf("[DEBUG] OTP: %s \n", otp.Passcode())
	}
}