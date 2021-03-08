package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Session")

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

type initDI struct {
	fx.In
	AppContext      *bootstrap.ApplicationContext
	SecRegistrar    security.Registrar `optional:"true"`
	SessionProps    security.SessionProperties
	ServerProps     web.ServerProperties
	ClientFactory   redis.ClientFactory
	MaxSessionsFunc GetMaximumSessions `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil {
		configurer := newSessionConfigurer(di.AppContext, di.SessionProps, di.ServerProps, di.ClientFactory, di.MaxSessionsFunc)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}