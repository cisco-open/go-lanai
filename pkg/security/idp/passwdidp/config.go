package passwdidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/formlogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"time"
)

// PasswordIdpSecurityConfigurer implements authserver.IdpSecurityConfigurer
type PasswordIdpSecurityConfigurer struct {
}

func NewPasswordIdpSecurityConfigurer() *PasswordIdpSecurityConfigurer {
	return &PasswordIdpSecurityConfigurer{}
}

func (c *PasswordIdpSecurityConfigurer) Configure(ws security.WebSecurity, config *authserver.Configuration) {
	// For Authorize endpoint
	handler := redirect.NewRedirectWithRelativePath("/error")
	condition := idp.RequestWithAuthenticationFlow(idp.InternalIdpForm, config.IdpManager)

	ws.AndCondition(condition).
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
			EnableMFA().
			LoginUrl("/login#/login").
			LoginErrorUrl("/login?error=true#/login").
			MfaUrl("/login/mfa#/otpverify").
			MfaErrorUrl("/login/mfa?error=true#/otpverify"),
		).
		With(errorhandling.New().
			AccessDeniedHandler(handler),
		).
		With(csrf.New().
			IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(config.Endpoints.Authorize.Location.Path)).
			IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(config.Endpoints.Logout)),
		).
		With(request_cache.New())
}

func debugPrintOTP(event passwd.MFAEvent, otp passwd.OTP, principal interface{}) {
	switch event {
	case passwd.MFAEventOtpCreate, passwd.MFAEventOtpRefresh:
		logger.Debugf("OTP: %s", otp.Passcode())
	}
}