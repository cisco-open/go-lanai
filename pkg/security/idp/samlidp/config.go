package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/samllogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

type SamlConfigurer struct {

}

func (c *SamlConfigurer) Configure(ws security.WebSecurity) {
	condition := matcher.RequestWithHost("saml.vms.com:8080") //TODO: this condition should be dynamic

	ws.Route(matcher.RouteWithPattern("/v2/authorize")).
		Condition(condition).
		With(samllogin.New()).
		With(session.New()).
		With(access.New().
			Request(
				matcher.RequestWithPattern("/saml/**"), //the two endpoints saml provides (/saml/metadata and /saml/sso) are public
			).PermitAll().
			Request(matcher.AnyRequest()).Authenticated(),
		)
}