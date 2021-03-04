package discovery

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module {
	Name: "service discovery",
	Precedence: bootstrap.ServiceDiscoveryPrecedence,
	Options: []fx.Option{
		fx.Provide(discovery.BindDiscoveryProperties, discovery.NewRegistration, discovery.NewCustomizers),
		fx.Invoke(setupServiceRegistration),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

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
