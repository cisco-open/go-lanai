package httpclient

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("HttpClient")

var Module = &bootstrap.Module{
	Name: "http-client",
	Precedence: bootstrap.HttpClientPrecedence,
	Options: []fx.Option{
		//fx.Provide(bindSecurityProperties),
		fx.Provide(provideHttpClient),
		//fx.Invoke(Test),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type clientDI struct {
	fx.In
	DiscClient discovery.Client
}
func provideHttpClient(di clientDI) Client {
	// TODO use properties
	return NewClient(di.DiscClient, func(opt *ClientOption) {
		opt.Verbose = true
	})
}