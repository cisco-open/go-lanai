package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
)

/*********************************
	Feature Impl
 *********************************/
//goland:noinspection GoNameStartsWithPackageName
type FormLoginFeature struct {
	formProcessCondition web.MWConditionMatcher
	successHandler       security.AuthenticationSuccessHandler
	failureHandler       security.AuthenticationErrorHandler
	loginUrl             string
	loginProcessUrl      string
	loginErrorUrl        string
	loginSuccessUrl      string
	usernameParam        string
	passwordParam        string
	rememberMeParam      string

	mfaEnabled    bool
	mfaUrl        string
	mfaVerifyUrl  string
	mfaRefreshUrl string
	mfaErrorUrl   string
	mfaSuccessUrl string
	otpParam      string
}

// Standard security.Feature entrypoint
func (f *FormLoginFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func (f *FormLoginFeature) LoginUrl(loginUrl string) *FormLoginFeature {
	f.loginUrl = loginUrl
	return f
}

func (f *FormLoginFeature) LoginProcessUrl(loginProcessUrl string) *FormLoginFeature {
	f.loginProcessUrl = loginProcessUrl
	return f
}

func (f *FormLoginFeature) FormProcessCondition(condition web.MWConditionMatcher) *FormLoginFeature {
	f.formProcessCondition = condition
	return f
}

func (f *FormLoginFeature) LoginSuccessUrl(loginSuccessUrl string) *FormLoginFeature {
	f.loginSuccessUrl = loginSuccessUrl
	return f
}

func (f *FormLoginFeature) LoginErrorUrl(loginErrorUrl string) *FormLoginFeature {
	f.loginErrorUrl = loginErrorUrl
	return f
}

func (f *FormLoginFeature) UsernameParameter(usernameParam string) *FormLoginFeature {
	f.usernameParam = usernameParam
	return f
}

func (f *FormLoginFeature) PasswordParameter(passwordParam string) *FormLoginFeature {
	f.passwordParam = passwordParam
	return f
}

func (f *FormLoginFeature) RememberMedParameter(rememberMeParam string) *FormLoginFeature {
	f.rememberMeParam = rememberMeParam
	return f
}

// SuccessHandler overrides LoginSuccessUrl
func (f *FormLoginFeature) SuccessHandler(successHandler security.AuthenticationSuccessHandler) *FormLoginFeature {
	f.successHandler = successHandler
	return f
}

// FailureHandler overrides LoginErrorUrl
func (f *FormLoginFeature) FailureHandler(failureHandler security.AuthenticationErrorHandler) *FormLoginFeature {
	f.failureHandler = failureHandler
	return f
}

func (f *FormLoginFeature) EnableMFA() *FormLoginFeature {
	f.mfaEnabled = true
	return f
}

func (f *FormLoginFeature) MfaUrl(mfaUrl string) *FormLoginFeature {
	f.mfaUrl = mfaUrl
	return f
}

func (f *FormLoginFeature) MfaVerifyUrl(mfaVerifyUrl string) *FormLoginFeature {
	f.mfaVerifyUrl = mfaVerifyUrl
	return f
}

func (f *FormLoginFeature) MfaRefreshUrl(mfaRefreshUrl string) *FormLoginFeature {
	f.mfaRefreshUrl = mfaRefreshUrl
	return f
}

func (f *FormLoginFeature) MfaSuccessUrl(mfaSuccessUrl string) *FormLoginFeature {
	f.mfaSuccessUrl = mfaSuccessUrl
	return f
}

func (f *FormLoginFeature) MfaErrorUrl(mfaErrorUrl string) *FormLoginFeature {
	f.mfaErrorUrl = mfaErrorUrl
	return f
}

func (f *FormLoginFeature) OtpParameter(otpParam string) *FormLoginFeature {
	f.otpParam = otpParam
	return f
}

/*********************************
	Constructors and Configure
 *********************************/
func Configure(ws security.WebSecurity) *FormLoginFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return  fc.Enable(feature).(*FormLoginFeature)
	}
	panic(fmt.Errorf("unable to configure form login: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *FormLoginFeature {
	return &FormLoginFeature{
		loginUrl:        "/login",
		loginProcessUrl: "/login",
		loginErrorUrl:   "/login?error=true",
		usernameParam:   "username",
		passwordParam:   "password",
		rememberMeParam: "remember-me",

		mfaUrl:        "/login/mfa",
		mfaVerifyUrl:  "/login/mfa",
		mfaRefreshUrl: "/login/mfa/refresh",
		mfaErrorUrl:   "/login/mfa?error=true",
		otpParam:      "otp",
	}
}