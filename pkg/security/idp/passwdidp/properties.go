package passwdidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

const (
	PwdAuthPropertiesPrefix = "security.idp.internal"
)

type PwdAuthProperties struct {
	Enabled                   bool                      `json:"enabled"`
	Domain                    string                    `json:"domain"`
	SessionExpiredRedirectUrl string                    `json:"session-expired-redirect-url"`
	Endpoints                 PwdAuthEndpointProperties `json:"endpoints"`
	MFA                       PwdAuthMfaProperties      `json:"mfa"`
	RememberMe                RememberMeProperties      `json:"remember-me"`
}

type PwdAuthEndpointProperties struct {
	FormLogin            string `json:"form-login"`
	FormLoginProcess     string `json:"form-login-process"`
	FormLoginError       string `json:"form-login-error"`
	OtpVerify            string `json:"otp-verify"`
	OtpVerifyProcess     string `json:"otp-verify-process"`
	OtpVerifyResend      string `json:"otp-verify-resend"`
	OtpVerifyError       string `json:"otp-verify-error"`
	ResetPasswordPageUrl string `json:"reset-password-page-url"`
}

type PwdAuthMfaProperties struct {
	Enabled        bool           `json:"enabled"`
	OtpLength      uint           `json:"otp-length"`
	OtpSecretSize  uint           `json:"otp-secret-size"`
	OtpTTL         utils.Duration `json:"otp-ttl"`
	OtpMaxAttempts uint           `json:"otp-max-attempts"`
	OtpResendLimit uint           `json:"otp-resend-limit"`
}

type RememberMeProperties struct {
	CookieDomain    string         `json:"cookie-domain"`
	UseSecureCookie bool           `json:"use-secure-cookie"`
	CookieValidity  utils.Duration `json:"cookie-validity"`
}

func NewPwdAuthProperties() *PwdAuthProperties {
	return &PwdAuthProperties{
		Domain: "localhost",
		Endpoints: PwdAuthEndpointProperties{
			FormLogin:            "/login",
			FormLoginProcess:     "/login",
			FormLoginError:       "/login?error=true",
			OtpVerify:            "/login/mfa",
			OtpVerifyProcess:     "/login/mfa",
			OtpVerifyResend:      "/login/mfa/refresh",
			OtpVerifyError:       "/login/mfa?error=true",
			ResetPasswordPageUrl: "/v2/todo",
		},
		MFA: PwdAuthMfaProperties{
			Enabled:        true,
			OtpLength:      6,
			OtpSecretSize:  10,
			OtpTTL:         utils.Duration(5 * time.Minute),
			OtpMaxAttempts: 5,
			OtpResendLimit: 5,
		},
		RememberMe: RememberMeProperties{
			CookieValidity: utils.Duration(2 * 7 * 24 * 60 * time.Minute),
		},
	}
}

func BindPwdAuthProperties(ctx *bootstrap.ApplicationContext) PwdAuthProperties {
	props := NewPwdAuthProperties()
	if err := ctx.Config().Bind(props, PwdAuthPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind PwdAuthProperties"))
	}
	return *props
}
