package swagger

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

type swaggerSecurityConfigurer struct {}

func (c *swaggerSecurityConfigurer) Configure(ws security.WebSecurity) {

	// DSL style example
	// for REST API
	ws.Route(matcher.RouteWithPattern("/v2/api-docs")).
		With(tokenauth.New()).
		With(access.New().
			Request(matcher.AnyRequest()).AllowIf(swaggerSpecAccessControl),
		).
		With(errorhandling.New())
}

func swaggerSpecAccessControl(auth security.Authentication) (decision bool, reason error) {
	//TODO: should differentiate between user auth and client auth,
	// should also check oauth2 scope?
	return access.Authenticated(auth)
}
