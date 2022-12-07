package swagger

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

type swaggerSecurityConfigurer struct{}

func (c *swaggerSecurityConfigurer) Configure(ws security.WebSecurity) {

	// DSL style example
	// for REST API
	ws.Route(matcher.RouteWithPattern("/v2/api-docs").Or(matcher.RouteWithPattern("/v3/api-docs"))).
		With(tokenauth.New()).
		With(access.New().
			Request(matcher.AnyRequest()).AllowIf(swaggerSpecAccessControl),
		).
		With(errorhandling.New())
}

func swaggerSpecAccessControl(auth security.Authentication) (decision bool, reason error) {
	oa, ok := auth.(oauth2.Authentication)
	if !ok {
		return false, security.NewInsufficientAuthError("expected token authentication")
	}

	if oa.UserAuthentication() == nil {
		return false, security.NewInsufficientAuthError("expected oauth user authentication")
	}

	if !(oa.OAuth2Request().Approved() && oa.OAuth2Request().Scopes().Has("read") && oa.OAuth2Request().Scopes().Has("write")) {
		return false, security.NewInsufficientAuthError("expected read and write scope")
	}

	//and must be authenticated
	return access.Authenticated(auth)
}
