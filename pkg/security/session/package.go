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
	Name:       "session",
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
	AppContext    *bootstrap.ApplicationContext
	SecRegistrar  security.Registrar `optional:"true"`
	SessionProps  security.SessionProperties
	ServerProps   web.ServerProperties         `optional:"true"`
	ClientFactory redis.ClientFactory          `optional:"true"`
	SettingReader security.GlobalSettingReader `optional:"true"`
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

	return NewRedisStore(redisClient, func(opt *StoreOption) {
		opt.SettingReader = di.SettingReader

		opt.Options.Path = path.Clean(di.SessionProps.Cookie.Path)
		opt.Options.Domain = di.SessionProps.Cookie.Domain
		opt.Options.MaxAge = di.SessionProps.Cookie.MaxAge
		opt.Options.Secure = di.SessionProps.Cookie.Secure
		opt.Options.HttpOnly = di.SessionProps.Cookie.HttpOnly
		opt.Options.SameSite = di.SessionProps.Cookie.SameSite()
		opt.Options.IdleTimeout = time.Duration(di.SessionProps.IdleTimeout)
		opt.Options.AbsoluteTimeout = time.Duration(di.SessionProps.AbsoluteTimeout)
	})
}

type initDI struct {
	fx.In
	AppContext            *bootstrap.ApplicationContext
	SecRegistrar          security.Registrar `optional:"true"`
	SessionProps          security.SessionProperties
	SessionStore          Store          `optional:"true"`
	SessionSettingService SettingService `optional:"true"`
}

func register(di initDI) {
	if di.SecRegistrar != nil && di.SessionStore != nil {
		configurer := newSessionConfigurer(di.SessionProps, di.SessionStore, di.SessionSettingService)
		di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
	}
}
