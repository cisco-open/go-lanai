package resserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var OAuth2AuthorizeModule = &bootstrap.Module{
	Name: "oauth2 authserver",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Provide(jwt.BindCryptoProperties),
		fx.Provide(ProvideResServerDI),
		fx.Invoke(ConfigureResourceServer),
	},
}

func Use() {
	security.Use()
	bootstrap.Register(OAuth2AuthorizeModule)
}

