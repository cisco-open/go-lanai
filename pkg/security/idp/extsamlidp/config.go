package extsamlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	samlsp "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/sp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
)

type Options func(opt *option)
type option struct {
	Properties *SamlAuthProperties
}

func WithProperties(props *SamlAuthProperties) Options {
	return func(opt *option) {
		opt.Properties = props
	}
}

// SamlIdpSecurityConfigurer implements authserver.IdpSecurityConfigurer
//goland:noinspection GoNameStartsWithPackageName
type SamlIdpSecurityConfigurer struct {
	props *SamlAuthProperties
}

func NewSamlIdpSecurityConfigurer(opts ...Options) *SamlIdpSecurityConfigurer {
	opt := option{
		Properties: NewSamlAuthProperties(),
	}
	for _, fn := range opts {
		fn(&opt)
	}
	return &SamlIdpSecurityConfigurer{
		props: opt.Properties,
	}
}

func (c *SamlIdpSecurityConfigurer) Configure(ws security.WebSecurity, config *authserver.Configuration) {
	// For Authorize endpoint
	condition := idp.RequestWithAuthenticationFlow(idp.ExternalIdpSAML, config.IdpManager)
	ws = ws.AndCondition(condition)

	if !c.props.Enabled {
		return
	}

	handler := redirect.NewRedirectWithURL(config.Endpoints.Error)
	ws.
		With(samlsp.New().
			Issuer(config.Issuer).
			ErrorPath(config.Endpoints.Error),
		).
		With(session.New().SettingService(config.SessionSettingService)).
		With(access.New().
			Request(matcher.AnyRequest()).Authenticated(),
		).
		With(errorhandling.New().
			AccessDeniedHandler(handler),
		)
}

func (c *SamlIdpSecurityConfigurer) ConfigureLogout(ws security.WebSecurity, config *authserver.Configuration) {
	if !c.props.Enabled {
		return
	}

	ws.With(samlsp.NewLogout().
		Issuer(config.Issuer).
		ErrorPath(config.Endpoints.Error),
	)
}
