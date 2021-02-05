package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/authconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/samllogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

// SamlIdpSecurityConfigurer implements authconfig.IdpSecurityConfigurer
type SamlIdpSecurityConfigurer struct {
	authFlowManager idp.AuthFlowManager
}

func NewSamlIdpSecurityConfigurer(authFlowManager idp.AuthFlowManager) *SamlIdpSecurityConfigurer {
	return &SamlIdpSecurityConfigurer{
		authFlowManager: authFlowManager,
	}
}

func (c *SamlIdpSecurityConfigurer) Configure(ws security.WebSecurity, config *authconfig.AuthorizationServerConfiguration) {
	handler := redirect.NewRedirectWithRelativePath("/error")
	condition := idp.RequestWithAuthenticationMethod(idp.ExternalIdpSAML, c.authFlowManager)

	ws.Condition(condition).
		With(samllogin.New()).
		With(session.New()).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(errorhandling.New().
			AccessDeniedHandler(handler),
		)
}