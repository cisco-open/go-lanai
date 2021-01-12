package logout

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"net/http"
)

/*********************************
	Feature Impl
 *********************************/
//goland:noinspection GoNameStartsWithPackageName
type LogoutHandler interface {
	//TODO
	HandleLogout(context.Context, *http.Request, http.ResponseWriter, security.Authentication)
}

//goland:noinspection GoNameStartsWithPackageName
type LogoutFeature struct {
	condition        web.RequestMatcher
	successHandler   security.AuthenticationSuccessHandler
	successUrl 	     string
	logoutHandlers   []LogoutHandler
	logoutUrl        string
}

// Standard security.Feature entrypoint
func (f *LogoutFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func (f *LogoutFeature) AddLogoutHandler(logoutHandler LogoutHandler) *LogoutFeature {
	f.logoutHandlers = append(f.logoutHandlers, logoutHandler)
	return f
}

func (f *LogoutFeature) Condition(condition web.RequestMatcher) *LogoutFeature {
	f.condition = condition
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

// SuccessHandler overrides SuccessUrl
func (f *LogoutFeature) SuccessHandler(successHandler security.AuthenticationSuccessHandler) *LogoutFeature {
	f.successHandler = successHandler
	return f
}

/*********************************
	Constructors and Configure
 *********************************/
func Configure(ws security.WebSecurity) *LogoutFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return  fc.Enable(feature).(*LogoutFeature)
	}
	panic(fmt.Errorf("unable to configure form login: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *LogoutFeature {
	return &LogoutFeature{
		successUrl:     "/login",
		logoutUrl:      "/logout",
		logoutHandlers: []LogoutHandler{},
	}
}