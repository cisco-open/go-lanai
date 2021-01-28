package passwdidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

type TokenEndpointSecurityConfigurer struct {
}

func (c *TokenEndpointSecurityConfigurer) Configure(ws security.WebSecurity) {
	// TODO
	// For Authorize endpoint
	handler := redirect.NewRedirectWithRelativePath("/error")
	condition := matcher.RequestWithHost("localhost:8080")

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
