package init

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"go.uber.org/fx"
)

var BasicAuthModule = &bootstrap.Module{
	Name: "basic auth",
	Precedence: security.MinSecurityPrecedence + 20,
	Options: []fx.Option{
		fx.Provide(newBasicAuthConfigurer),
		fx.Invoke(setup),
	},
}

func init() {
	bootstrap.Register(BasicAuthModule)
}

func setup(init security.Initializer, c *BasicAuthConfigurer) {
	init.(security.FeatureRegistrar).RegisterFeatureConfigurer(BasicAuthConfigurerType, c)
}
