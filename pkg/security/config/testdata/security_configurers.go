package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

type ErrorPageSecurityConfigurer struct {}

func (c *ErrorPageSecurityConfigurer) Configure(ws security.WebSecurity) {

	ws.Route(matcher.RouteWithPattern("/error")).
		With(session.New()).
		With(access.New().
			Request(matcher.AnyRequest()).PermitAll(),
		)
}
