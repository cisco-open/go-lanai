package formlogin

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

/*********************************
	Feature Impl
 *********************************/
//goland:noinspection GoNameStartsWithPackageName
type FormLoginFeature struct {
	successHandler  security.AuthenticationSuccessHandler
	failureHandler  security.AuthenticationErrorHandler
	loginUrl        string
	loginProcessUrl string
	loginErrorUrl   string
	loginSuccessUrl string
	logoutUrl       string
	usernameParam   string
	passwordParam   string
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


func (f *FormLoginFeature) LoginSuccessUrl(loginSuccessUrl string) *FormLoginFeature {
	f.loginSuccessUrl = loginSuccessUrl
	return f
}

func (f *FormLoginFeature) LoginErrorUrl(loginErrorUrl string) *FormLoginFeature {
	f.loginErrorUrl = loginErrorUrl
	return f
}

func (f *FormLoginFeature) LogoutUrl(logoutUrl string) *FormLoginFeature {
	f.logoutUrl = logoutUrl
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
		logoutUrl:       "/logout",
		usernameParam:   "username",
		passwordParam:   "password",
	}
}