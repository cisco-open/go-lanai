package logout

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"net/http"
)

/*********************************
	Feature Impl
 *********************************/
//goland:noinspection GoNameStartsWithPackageName
type LogoutHandler interface {
	HandleLogout(context.Context, *http.Request, http.ResponseWriter, security.Authentication) error
}

//goland:noinspection GoNameStartsWithPackageName
type LogoutFeature struct {
	successHandler security.AuthenticationSuccessHandler
	errorHandler   security.AuthenticationErrorHandler
	successUrl     string
	errorUrl       string
	logoutHandlers []LogoutHandler
	logoutUrl      string
}

// Standard security.Feature entrypoint
func (f *LogoutFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

// LogoutHandlers override default handler
func (f *LogoutFeature) LogoutHandlers(logoutHandlers ...LogoutHandler) *LogoutFeature {
	f.logoutHandlers = logoutHandlers
	return f
}

func (f *LogoutFeature) AddLogoutHandler(logoutHandler LogoutHandler) *LogoutFeature {
	f.logoutHandlers = append([]LogoutHandler{logoutHandler}, f.logoutHandlers...)
	return f
}

func (f *LogoutFeature) LogoutUrl(logoutUrl string) *LogoutFeature {
	f.logoutUrl = logoutUrl
	return f
}

func (f *LogoutFeature) SuccessUrl(successUrl string) *LogoutFeature {
	f.successUrl = successUrl
	return f
}

func (f *LogoutFeature) ErrorUrl(errorUrl string) *LogoutFeature {
	f.errorUrl = errorUrl
	return f
}

// SuccessHandler overrides SuccessUrl
func (f *LogoutFeature) SuccessHandler(successHandler security.AuthenticationSuccessHandler) *LogoutFeature {
	f.successHandler = successHandler
	return f
}

// ErrorHandler overrides ErrorUrl
func (f *LogoutFeature) ErrorHandler(errorHandler security.AuthenticationErrorHandler) *LogoutFeature {
	f.errorHandler = errorHandler
	return f
}

/*********************************
	Constructors and Configure
 *********************************/
func Configure(ws security.WebSecurity) *LogoutFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*LogoutFeature)
	}
	panic(fmt.Errorf("unable to configure form login: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *LogoutFeature {
	return &LogoutFeature{
		successUrl: "/login",
		logoutUrl:  "/logout",
		logoutHandlers: []LogoutHandler{
			DefaultLogoutHanlder{},
		},
	}
}

type DefaultLogoutHanlder struct{}

func (h DefaultLogoutHanlder) HandleLogout(ctx context.Context, _ *http.Request, _ http.ResponseWriter, _ security.Authentication) error {
	security.Clear(ctx)
	return nil
}
