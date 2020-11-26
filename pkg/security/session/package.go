package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var SessionModule = &bootstrap.Module{
	Name: "session",
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		fx.Provide(security.BindSessionProperties),
		fx.Invoke(register),
	},
}


func init() {
	bootstrap.Register(SessionModule)
	security.GobRegister()
	passwd.GobRegister()
}


func register(init security.Registrar, sessionProps security.SessionProperties, serverProps web.ServerProperties, connection *redis.Connection) {
	configurer := newSessionConfigurer(sessionProps, serverProps, connection)
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}
