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
	gob.Register((*CachedRequest)(nil))
	gob.Register([]interface{}{})
}

func register(init security.Registrar, sessionProps security.SessionProperties, serverProps web.ServerProperties, client redis.Client) {
	configurer := newSessionConfigurer(sessionProps, serverProps, client)
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}