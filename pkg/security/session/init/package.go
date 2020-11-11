package init

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var SessionModule = &bootstrap.Module{
	Name: "session",
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		fx.Provide(security.BindSessionProperties, newSessionConfigurer),
		fx.Invoke(register),
	},
}


func init() {
	bootstrap.Register(SessionModule)
	security.GobRegister()
	passwd.GobRegister()
}


func register(init security.Registrar, sessionProps security.SessionProperties, serverProps web.ServerProperties) {
	configurer := newSessionConfigurer(sessionProps, serverProps)
	init.(security.FeatureRegistrar).RegisterFeature(SessionFeatureId, configurer)
}
