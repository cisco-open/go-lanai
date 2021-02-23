package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "session",
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		fx.Provide(security.BindSessionProperties),
		fx.Invoke(register),
	},
}


func init() {
	bootstrap.Register(Module)

	GobRegister()
	security.GobRegister()
	passwd.GobRegister()
}

func GobRegister() {
	gob.Register([]interface{}{})
}

type registerParams struct {
	fx.In
	Init security.Registrar
	SessionProps security.SessionProperties
	ServerProps web.ServerProperties
	ClientFactory redis.ClientFactory
	MaxSessionsFunc GetMaximumSessions `optional:"true"`
}

func register(di registerParams) {
	configurer := newSessionConfigurer(di.SessionProps, di.ServerProps, di.ClientFactory, di.MaxSessionsFunc)
	di.Init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}