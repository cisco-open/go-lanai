package unknownIdp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"fmt"
)

type NoIdpSecurityConfigurer struct {
}

func NewNoIdpSecurityConfigurer() *NoIdpSecurityConfigurer {
	return &NoIdpSecurityConfigurer{}
}

func (c *NoIdpSecurityConfigurer) Configure(ws security.WebSecurity, config *authserver.Configuration) {
	// For Authorize endpoint
	handler := redirect.NewRedirectWithURL(config.Endpoints.Error)
	condition := idp.RequestWithAuthenticationFlow(idp.UnknownIdp, config.IdpManager)

	ws.AndCondition(condition).
		With(session.New().SettingService(config.SessionSettingService)).
		With(access.New().
			Request(matcher.AnyRequest()).
			AllowIf(authenticatedWithMessage("Identity provider is not configured for this sub-domain")),
		).
		With(errorhandling.New().
			AccessDeniedHandler(handler),
		)
}

func authenticatedWithMessage(format string, v...interface{}) access.ControlFunc {
	return func(auth security.Authentication) (decision bool, reason error) {
		if auth.State() >= security.StateAuthenticated {
			return true, nil
		} else {
			return false, security.NewAccessDeniedError(fmt.Sprintf(format, v...))
		}
	}
}