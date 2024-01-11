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
			discovery.NewCustomizers,
			provideRegistration,
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

type regDI struct {
	fx.In
	AppContext          *bootstrap.ApplicationContext
	DiscoveryProperties discovery.DiscoveryProperties
}

func provideRegistration(di regDI) *api.AgentServiceRegistration {
	return discovery.NewRegistration(discovery.RegistrationWithProperties(di.AppContext, di.DiscoveryProperties))
}

func provideDiscoveryClient(ctx *bootstrap.ApplicationContext, conn *consul.Connection, props discovery.DiscoveryProperties) discovery.Client {
	return discovery.NewConsulDiscoveryClient(ctx, conn, func(opt *discovery.ClientConfig) {
		opt.DefaultSelector = discovery.InstanceWithProperties(&props.DefaultSelector)
	})
}

func setupServiceRegistration(lc fx.Lifecycle,
	connection *consul.Connection, registration *api.AgentServiceRegistration, customizers *discovery.Customizers) {

	//because we are the lowest precendence, we execute when every thing is ready
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			customizers.Apply(ctx, registration)
			_ = discovery.Register(ctx, connection, registration)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			customizers.Apply(ctx, registration)
			_ = discovery.Deregister(ctx, connection, registration)
			return nil
		},
	})
}
