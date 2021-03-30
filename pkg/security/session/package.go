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
	"path"
	"time"
)

var logger = log.New("SEC.Session")

var Module = &bootstrap.Module{
	Name: "session",
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		fx.Provide(security.BindSessionProperties),
		fx.Provide(provideSessionStore),
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

type storeDI struct {
	fx.In
	AppContext      *bootstrap.ApplicationContext
	SecRegistrar    security.Registrar `optional:"true"`
	SessionProps    security.SessionProperties
	ServerProps     web.ServerProperties `optional:"true"`
	ClientFactory   redis.ClientFactory  `optional:"true"`
	MaxSessionsFunc GetMaximumSessions   `optional:"true"`
}

func provideSessionStore(di storeDI) Store {
	if di.SecRegistrar == nil || di.ClientFactory == nil {
		return nil
	}
	redisClient, e := di.ClientFactory.New(di.AppContext, func(opt *redis.ClientOption) {
		opt.DbIndex = di.SessionProps.DbIndex
	})
	if e != nil {
		panic(e)
	}

	configureOptions := func(options *Options) {
		options.Path = path.Clean("/" + di.ServerProps.ContextPath)
		options.Domain = di.SessionProps.Cookie.Domain
		options.MaxAge = di.SessionProps.Cookie.MaxAge
		options.Secure = di.SessionProps.Cookie.Secure
		options.HttpOnly = di.SessionProps.Cookie.HttpOnly
		options.SameSite = di.SessionProps.Cookie.SameSite()
		options.IdleTimeout = time.Duration(di.SessionProps.IdleTimeout)
		options.AbsoluteTimeout = time.Duration(di.SessionProps.AbsoluteTimeout)
	}

	return NewRedisStore(redisClient, configureOptions)
}

type initDI struct {
	fx.In
	AppContext      *bootstrap.ApplicationContext
	SecRegistrar    security.Registrar `optional:"true"`
	SessionProps    security.SessionProperties
	SessionStore    Store              `optional:"true"`
	MaxSessionsFunc GetMaximumSessions `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil && di.SessionStore != nil {
		configurer := newSessionConfigurer(di.SessionProps, di.SessionStore, di.MaxSessionsFunc)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}