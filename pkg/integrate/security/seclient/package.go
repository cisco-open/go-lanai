package seclient

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Client")

var Module = &bootstrap.Module{
	Name: "auth-client",
	Precedence: bootstrap.SecurityIntegrationPrecedence,
	Options: []fx.Option{
		//fx.Provide(bindHttpClientProperties),
		fx.Provide(provideAuthClient),
	},
}

func Use() {
	httpclient.Use()
	bootstrap.Register(Module)
}

type clientDI struct {
	fx.In
	HttpClient  httpclient.Client
}

func provideAuthClient(di clientDI) AuthenticationClient {
	return NewRemoteAuthClient(func(opt *AuthClientOption) {
		// TODO use properties
		opt.Client = di.HttpClient
		opt.ClientId = "nfv-service"
		opt.ClientSecret = "nfv-service-secret"
		opt.ServiceName = "europa"
	})
}

