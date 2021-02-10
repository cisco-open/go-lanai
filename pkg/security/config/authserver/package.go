package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

//goland:noinspection GoNameStartsWithPackageName
var OAuth2AuthorizeModule = &bootstrap.Module{
	Name: "oauth2 authserver",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Provide(bindProperties),
		fx.Invoke(ConfigureAuthorizationServer),
	},
}

func init() {
	bootstrap.Register(OAuth2AuthorizeModule)
}

func Use() {

}

func bindProperties(ctx *bootstrap.ApplicationContext) security.SamlProperties {
	props := security.NewSamlProperties()
	if err := ctx.Config().Bind(props, security.SamlPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SamlProperties"))
	}
	return *props
}

