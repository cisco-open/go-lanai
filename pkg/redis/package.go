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
		fx.Provide(NewClientFactory),
		fx.Provide(newDefaultClient),
		fx.Invoke(registerHealth),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

type clientDI struct {
	fx.In
	AppCtx             *bootstrap.ApplicationContext
	Factory            ClientFactory
	Properties         RedisProperties
	TLSProviderFactory *tlsconfig.ProviderFactory `optional:"true"`
}

func newDefaultClient(di clientDI) Client {
	c, e := di.Factory.New(di.AppCtx, func(opt *ClientOption) {
		opt.DbIndex = di.Properties.DB
		opt.TLSProviderFactory = di.TLSProviderFactory
	})

	if e != nil {
		panic(e)
	}
	return c
}
