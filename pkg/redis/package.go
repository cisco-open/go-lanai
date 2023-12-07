package redis

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	"go.uber.org/fx"
)

var logger = log.New("Redis")

var Module = &bootstrap.Module{
	Precedence: bootstrap.RedisPrecedence,
	Options: []fx.Option{
		fx.Provide(BindRedisProperties),
		fx.Provide(provideClientFactory),
		fx.Provide(provideDefaultClient),
		fx.Invoke(registerHealth),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

type factoryDI struct {
	fx.In
	Props RedisProperties
	CertManager tlsconfig.Manager `optional:"true"`
}

func provideClientFactory(di factoryDI) ClientFactory {
	return NewClientFactory(func(opt *FactoryOption) {
		opt.Properties = di.Props
		opt.TLSCertsManager = di.CertManager
	})
}

type clientDI struct {
	fx.In
	AppCtx             *bootstrap.ApplicationContext
	Factory            ClientFactory
	Properties         RedisProperties
}

func provideDefaultClient(di clientDI) Client {
	c, e := di.Factory.New(di.AppCtx, func(opt *ClientOption) {
		opt.DbIndex = di.Properties.DB
	})

	if e != nil {
		panic(e)
	}
	return c
}
