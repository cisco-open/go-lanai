package redis

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: bootstrap.RedisPrecedence,
	Options: []fx.Option{
		fx.Provide(BindSessionProperties),
		fx.Provide(NewClientFactory),
		fx.Provide(newDefaultClient),
		fx.Invoke(registerHealth),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func newDefaultClient(f ClientFactory, p ConnectionProperties) Client {
	c, e := f.New(func(opt *ClientOption) {
		opt.DbIndex = p.DB
	})

	if e != nil {
		panic(e)
	}
	return c
}
