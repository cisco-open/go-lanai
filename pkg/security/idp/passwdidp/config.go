package passwdidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

type TokenEndpointSecurityConfigurer struct {
	authFlowManager idp.AuthFlowManager
}

func (c *TokenEndpointSecurityConfigurer) Configure(ws security.WebSecurity) {
	// TODO
	// For Authorize endpoint
	handler := redirect.NewRedirectWithRelativePath("/error")
	condition := idp.RequestWithAuthenticationMethod(idp.InternalIdpForm, c.authFlowManager)

	ws.Route(matcher.RouteWithPattern("/v2/authorize")).
		Condition(condition).
		With(passwd.New()).
		//With(token.NewTokenEndpoint()).
		//With(access.New().
		//	RequestDetails(matcher.AnyRequest()).Authenticated(),
		//).
		With(errorhandling.New().
			AuthenticationEntryPoint(handler).
			AccessDeniedHandler(handler),
		)
}
