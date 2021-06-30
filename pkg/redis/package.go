package redis

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
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

func newDefaultClient(ctx *bootstrap.ApplicationContext, f ClientFactory, p RedisProperties) Client {
	c, e := f.New(ctx, func(opt *ClientOption) {
		opt.DbIndex = p.DB
	})

	if e != nil {
		panic(e)
	}
	return c
}
