package authserver

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/timeoutsupport"
	saml_auth "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/saml_sso"
	th_loader "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy/loader"
	"embed"
	"go.uber.org/fx"
)

//go:embed defaults-authserver.yml
var defaultConfigFS embed.FS

//goland:noinspection GoNameStartsWithPackageName
var OAuth2AuthorizeModule = &bootstrap.Module{
	Name: "oauth2 authserver",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
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
	bootstrap.Register(OAuth2AuthorizeModule)
	timeoutsupport.Use()
	// Note: External SAML IDP support (samllogin package) is enabled as part of samlidp
}

