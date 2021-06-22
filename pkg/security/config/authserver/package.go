package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/timeoutsupport"
	saml_auth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin"
	th_loader "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy/loader"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var OAuth2AuthorizeModule = &bootstrap.Module{
	Name: "oauth2 authserver",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Provide(BindAuthServerProperties),
		fx.Provide(ProvideAuthServerDI),
		fx.Provide(provide),
		fx.Invoke(ConfigureAuthorizationServer),
	},
}

func Use() {
	security.Use()
	th_loader.Use()
	saml_auth.Use() // saml_auth enables SAML SSO
	samllogin.Use() // samllogin enables External SAML IDP support
	bootstrap.Register(OAuth2AuthorizeModule)
	timeoutsupport.Use()
}

