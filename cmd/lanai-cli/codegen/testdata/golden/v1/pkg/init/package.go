package serviceinit

import (
	actuator "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/init"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	consul "cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul/init"
	discovery "cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/swagger"
	tracing "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing/init"
	vault "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault/init"
	web "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
	"cto-github.cisco.com/NFV-BU/test-service/pkg/controller"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "testservice",
	Precedence: bootstrap.AnonymousModulePrecedence,
	Options: []fx.Option{
		fx.Provide(newResServerConfigurer),
		fx.Invoke(configureSecurity),
	},
}

// Use initialize components needed in this service
func Use() {
	// basic modules
	appconfig.Use()
	consul.Use()
	vault.Use()
	redis.Use()
	tracing.Use()

	// web related
	web.Use()
	actuator.Use()
	swagger.Use()

	// data related
	//data.Use()
	//cockroach.Use()

	// service-to-service integration related
	discovery.Use()
	//httpclient.Use()
	//scope.Use()
	//kafka.Use()

	// security related modules
	security.Use()
	resserver.Use()
	//opainit.Use()

	// testservice
	bootstrap.Register(Module)
	bootstrap.Register(controller.Module)
	for _, m := range controller.SubModules {
		bootstrap.Register(m)
	}
}
