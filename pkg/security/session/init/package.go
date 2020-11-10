package init

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"go.uber.org/fx"
)

var SessionModule = &bootstrap.Module{
	Name: "session",
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		fx.Provide(security.BindSessionProperties, newSessionConfigurer),
		fx.Invoke(setup),
	},
}


func init() {
	bootstrap.Register(SessionModule)
	security.GobRegister()
	passwd.GobRegister()
}

func setup(init security.Registrar, c *SessionConfigurer) {
	init.(security.FeatureRegistrar).RegisterFeatureConfigurer(SessionConfigurerType, c)
}
