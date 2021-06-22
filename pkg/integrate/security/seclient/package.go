package seclient

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	securityint "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("SEC.Client")

var Module = &bootstrap.Module{
	Name: "auth-client",
	Precedence: bootstrap.SecurityIntegrationPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(securityint.DefaultConfigFS),
		fx.Provide(securityint.BindSecurityIntegrationProperties),
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
	Properties securityint.SecurityIntegrationProperties
}

func provideAuthClient(di clientDI) AuthenticationClient {
	return NewRemoteAuthClient(func(opt *AuthClientOption) {
		opt.Client = di.HttpClient
		opt.ServiceName = di.Properties.ServiceName
		opt.ClientId = di.Properties.Client.ClientId
		opt.ClientSecret = di.Properties.Client.ClientSecret
		opt.BaseUrl = di.Properties.Endpoints.BaseUrl
		opt.PwdLoginPath = di.Properties.Endpoints.PasswordLogin
		opt.SwitchContextPath = di.Properties.Endpoints.SwitchContext
	})
}

