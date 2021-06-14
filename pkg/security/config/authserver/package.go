package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/timeoutsupport"
	th_loader "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy/loader"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var OAuth2AuthorizeModule = &bootstrap.Module{
	Name: "oauth2 authserver",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Provide(BindAuthServerProperties),
		fx.Provide(NewConfiguration),
		fx.Provide(provide),
		fx.Invoke(ConfigureAuthorizationServer),
	},
}

func Use() {
	security.Use()
	th_loader.Use()
	bootstrap.Register(OAuth2AuthorizeModule)
	timeoutsupport.Use()
}

