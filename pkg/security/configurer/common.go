package configurer

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	session "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/init"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		//fx.Provide(security.BindSessionProperties),
		fx.Invoke(initialize),
	},
}

type SecurityInitializer struct {
	fx.In
	sessionConfigurer *session.SessionConfigurer
}

func initialize(initializer *SecurityInitializer) {

}


