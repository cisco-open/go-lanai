package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

// SamlIdpSecurityConfigurer implements authserver.IdpSecurityConfigurer
type SamlIdpSecurityConfigurer struct {
}

func NewSamlIdpSecurityConfigurer() *SamlIdpSecurityConfigurer {
	return &SamlIdpSecurityConfigurer{}
}

func (c *SamlIdpSecurityConfigurer) Configure(ws security.WebSecurity, config *authserver.Configuration) {
	handler := redirect.NewRedirectWithRelativePath("/error")
	condition := idp.RequestWithAuthenticationFlow(idp.ExternalIdpSAML, config.IdpManager)

	ws.AndCondition(condition).
		With(samllogin.New().
			Issuer(config.Issuer),
		).
		With(session.New()).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(errorhandling.New().
			AccessDeniedHandler(handler),
		)
}