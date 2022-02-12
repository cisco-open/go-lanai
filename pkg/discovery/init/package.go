package discovery

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"embed"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
)

//go:embed defaults-discovery.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module {
	Name: "service discovery",
	Precedence: bootstrap.ServiceDiscoveryPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(discovery.BindDiscoveryProperties,
			discovery.NewRegistration,
			discovery.NewCustomizers,
			provideDiscoveryClient),
		fx.Invoke(setupServiceRegistration),
	},
}

func init() {
	bootstrap.Register(Module)
}

func Use() {
	// does nothing. Allow service to include this module in main()
}

func provideDiscoveryClient(ctx *bootstrap.ApplicationContext, conn *consul.Connection) discovery.Client {
	return discovery.NewConsulDiscoveryClient(ctx, conn)
}

func setupServiceRegistration(lc fx.Lifecycle,
	connection *consul.Connection, registration *api.AgentServiceRegistration, customizers *discovery.Customizers) {

	//because we are the lowest precendence, we execute when every thing is ready
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			customizers.Apply(ctx, registration)
			discovery.Register(ctx, connection, registration)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			customizers.Apply(ctx, registration)
			discovery.Deregister(ctx, connection, registration)
			return nil
		},
	})
}
