package servicedisc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	sdcustomizer "cto-github.cisco.com/NFV-BU/go-lanai/pkg/servicedisc/customizer"
	netutil "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/net"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
)

var logger = log.New("service discovery")

var Module = &bootstrap.Module {
	Name: "service discovery",
	Precedence: bootstrap.ServiceDiscoveryPrecedence,
	Options: []fx.Option{
		fx.Provide(newDiscoveryProperties, newRegistration, newCustomizerRegistrar),
		fx.Invoke(setupServiceRegistration),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

func newDiscoveryProperties(appConfig *appconfig.ApplicationConfig, serverProps web.ServerProperties) DiscoveryProperties {
	ipAddress, _ := netutil.GetIp("")
	p := DiscoveryProperties{
		IpAddress: ipAddress,
		Port: serverProps.Port,
		Scheme: "http",
		HealthCheckInterval: "15s",
		HealthCheckPath: fmt.Sprintf("%s%s", serverProps.ContextPath, "/admin/health"),
	}
	appConfig.Bind(p, discoveryPropertiesPrefix)
	return p
}

func newCustomizerRegistrar(appContext *bootstrap.ApplicationContext) *sdcustomizer.Registrar{
	r := sdcustomizer.NewRegistrar()
	r.Add(NewDefaultCustomizer(appContext))
	return r
}

func setupServiceRegistration(lc fx.Lifecycle,
	connection *consul.Connection, registration *api.AgentServiceRegistration, customizers *sdcustomizer.Registrar) {
	for _, c := range customizers.Customizers {
		c.Customize(registration)
	}

	//because we are the lowest precendence, we execute when every thing is ready
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			registrar := kitconsul.NewRegistrar(kitconsul.NewClient(connection.Client()), registration , logger.WithContext(ctx))
			registrar.Register()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			registrar := kitconsul.NewRegistrar(kitconsul.NewClient(connection.Client()), registration , logger.WithContext(ctx))
			registrar.Deregister()
			return nil
		},
	})
}
