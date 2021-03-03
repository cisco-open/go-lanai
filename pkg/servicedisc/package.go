package servicedisc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	sdcustomizer "cto-github.cisco.com/NFV-BU/go-lanai/pkg/servicedisc/customizer"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	netutil "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/net"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	kitlog "github.com/go-kit/kit/log"
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
)

var logger = log.GetNamedLogger("service discovery")

var Module = &bootstrap.Module {
	Name: "service discovery",
	Precedence: bootstrap.ServiceDiscoveryPrecedence,
	Options: []fx.Option{
		fx.Provide(newApplicationProperties, newDiscoveryProperties, newRegistration, sdcustomizer.NewRegistrar),
		fx.Invoke(setupServiceRegistration),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

//TODO: move to app config or bootstrap
func newApplicationProperties(appConfig *appconfig.ApplicationConfig) *ApplicationProperties {
	p := &ApplicationProperties{}
	appConfig.Bind(p, applicationPropertiesPrefix)
	return p
}

func newDiscoveryProperties(appConfig *appconfig.ApplicationConfig, serverProps web.ServerProperties) *DiscoveryProperties {
	ipAddress, _ := netutil.GetIp("")
	p := &DiscoveryProperties{
		IpAddress: ipAddress,
		Port: serverProps.Port,
		Scheme: "http",
		HealthCheckInterval: "15s",
		HealthCheckPath: fmt.Sprintf("%s%s", serverProps.ContextPath, "/admin/health"),
	}
	appConfig.Bind(p, discoveryPropertiesPrefix)
	return p
}

//TODO: compare tags
func newRegistration(appProps *ApplicationProperties, discoveryProperties *DiscoveryProperties) *api.AgentServiceRegistration {
	registration := &api.AgentServiceRegistration{
		Kind: api.ServiceKindTypical,
		ID:   fmt.Sprintf("%s-%d-%x", appProps.Name, discoveryProperties.Port, cryptoutils.RandomBytes(5)),
		Name: appProps.Name,
		Tags: discoveryProperties.Tags,
		Port: discoveryProperties.Port,
		Address: discoveryProperties.IpAddress,
		Check: &api.AgentServiceCheck{
			HTTP: fmt.Sprintf("%s://%s:%d%s", discoveryProperties.Scheme, discoveryProperties.IpAddress, discoveryProperties.Port, discoveryProperties.HealthCheckPath),
			Interval: discoveryProperties.HealthCheckInterval,
			DeregisterCriticalServiceAfter: discoveryProperties.HealthCheckCriticalTimeout},
		}
	return registration
}

func setupServiceRegistration(lc fx.Lifecycle,
	connection *consul.Connection, registration *api.AgentServiceRegistration, customizers *sdcustomizer.Registrar) {
	for _, c := range customizers.Customizers {
		c.Customize(registration)
	}

	//TODO: provide our own logger
	//because we are the lowest precendence, we execute when every thing is ready
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			//TODO: logger with context
			registrar := kitconsul.NewRegistrar(kitconsul.NewClient(connection.Client()), registration ,kitlog.NewNopLogger())
			registrar.Register()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			//TODO: logger with context
			registrar := kitconsul.NewRegistrar(kitconsul.NewClient(connection.Client()), registration ,kitlog.NewNopLogger())
			registrar.Deregister()
			return nil
		},
	})
}
