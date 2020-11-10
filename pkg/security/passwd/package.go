package passwd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

var PasswordAuthModule = &bootstrap.Module{
	Name: "passwd authenticator",
	Precedence: security.MinSecurityPrecedence + 30,
	Options: []fx.Option{
		fx.Provide(newPasswordAuthConfigurer),
		fx.Invoke(setup),
	},
}

func init() {
	bootstrap.Register(PasswordAuthModule)
}

func setup(init security.Registrar, c *PasswordAuthConfigurer) {
	init.(security.FeatureRegistrar).RegisterFeatureConfigurer(PasswordAuthConfigurerType, c)
}