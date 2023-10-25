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
	bootstrap.Register(tlsconfig.Module)
	bootstrap.Register(Module)
}

func newDefaultClient(ctx *bootstrap.ApplicationContext, f ClientFactory, p RedisProperties, t *tlsconfig.ProviderFactory) Client {
	c, e := f.New(ctx, func(opt *ClientOption) {
		opt.DbIndex = p.DB
		opt.TlsProviderFactory = t
	})

	if e != nil {
		panic(e)
	}
	return c
}
