package redis

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: bootstrap.LowestPrecedence,
	Options: []fx.Option{
		fx.Provide(newConnectionProperties),
		fx.Provide(newClient),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func newConnectionProperties(ac *bootstrap.ApplicationContext) *ConnectionProperties {
	r := &ConnectionProperties{}
	ac.Config().Bind(r, ConfigRootRedisConnection)
	return r
}

func newClient(p *ConnectionProperties) Client {

	opts, err := GetUniversalOptions(p)

	if err != nil {
		panic(errors.Wrap(err, "Invalid redis configuration"))
	}

	rdb := redis.NewUniversalClient(opts)

	c := &client {
		UniversalClient: rdb,
	}

	return c
}
