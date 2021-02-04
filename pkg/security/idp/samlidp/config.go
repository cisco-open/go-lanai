package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/samllogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

type SamlConfigurer struct {
	authFlowManager idp.AuthFlowManager
}

func (c *SamlConfigurer) Configure(ws security.WebSecurity) {
	condition := idp.RequestWithAuthenticationMethod(idp.ExternalIdpSAML, c.authFlowManager)

	ws.Route(matcher.RouteWithPattern("/v2/authorize")).
		Condition(condition).
		With(samllogin.New()).
		With(session.New()).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		)
}