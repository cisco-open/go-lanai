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

type Options func(opt *option)
type option struct {
	ErrorPath string
}

func WithErrorPath(path string) Options {
	return func(opt *option) {
		opt.ErrorPath = path
	}
}

// SamlIdpSecurityConfigurer implements authserver.IdpSecurityConfigurer
//goland:noinspection GoNameStartsWithPackageName
type SamlIdpSecurityConfigurer struct {
	errorPath string
}

func NewSamlIdpSecurityConfigurer(opts ...Options) *SamlIdpSecurityConfigurer {
	opt := option {
		ErrorPath: "/error",
	}
	for _, fn := range opts {
		fn(&opt)
	}
	return &SamlIdpSecurityConfigurer{
		errorPath: opt.ErrorPath,
	}
}

func (c *SamlIdpSecurityConfigurer) Configure(ws security.WebSecurity, config *authserver.Configuration) {
	handler := redirect.NewRedirectWithRelativePath(c.errorPath)
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