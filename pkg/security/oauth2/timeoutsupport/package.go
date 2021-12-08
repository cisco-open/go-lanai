package timeoutsupport

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"go.uber.org/fx"
)

//var logger = log.New("SEC.Timeout")

var Module = &bootstrap.Module{
	Name: "timeout",
	Precedence: security.MinSecurityPrecedence + 10, //same as session. since this package doesn't invoke anything, the precedence has no real effect
	Options: []fx.Option{
		fx.Provide(security.BindTimeoutSupportProperties),
		fx.Provide(provideTimeoutSupport),
	},
}

func provideTimeoutSupport(ctx *bootstrap.ApplicationContext, cf redis.ClientFactory, prop security.TimeoutSupportProperties) oauth2.TimeoutApplier {
	client, err := cf.New(ctx, func(opt *redis.ClientOption) {
		opt.DbIndex = prop.DbIndex
	})

	if err != nil {
		panic(err)
	}

	support := NewRedisTimeoutApplier(client)
	return support
}

func Use() {
	bootstrap.Register(Module)
}