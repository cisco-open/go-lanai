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
	handler := redirect.NewRedirectWithRelativePath(config.Endpoints.Error)
	condition := idp.RequestWithAuthenticationFlow(idp.UnknownIdp, config.IdpManager)

	ws.AndCondition(condition).
		With(session.New().SettingService(config.SessionSettingService)).
		With(access.New().
			Request(matcher.AnyRequest()).
			AllowIf(alwaysDenyWithMessage("Identity provider is not configured for this sub-domain")),
		).
		With(errorhandling.New().
			AccessDeniedHandler(handler),
		)
}

func alwaysDenyWithMessage(format string, v...interface{}) access.ControlFunc {
	return func(_ security.Authentication) (decision bool, reason error) {
		return false, security.NewAccessDeniedError(fmt.Sprintf(format, v...))
	}
}