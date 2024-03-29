package serviceinit

import (
	actuator "github.com/cisco-open/go-lanai/pkg/actuator/init"
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	consul "github.com/cisco-open/go-lanai/pkg/consul/init"
	"github.com/cisco-open/go-lanai/pkg/discovery/consulsd"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/swagger"
	tracing "github.com/cisco-open/go-lanai/pkg/tracing/init"
	vault "github.com/cisco-open/go-lanai/pkg/vault/init"
	web "github.com/cisco-open/go-lanai/pkg/web/init"
	"github.com/cisco-open/test-service/pkg/controller"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "testservice",
	Precedence: bootstrap.AnonymousModulePrecedence,
	Options:    []fx.Option{},
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
	consulsd.Use()
	//httpclient.Use()
	//scope.Use()
	//kafka.Use()

	// security related modules
	//security.Use()
	//resserver.Use()
	//opainit.Use()

	// testservice
	bootstrap.Register(Module)
	bootstrap.Register(controller.Module)
	for _, m := range controller.SubModules {
		bootstrap.Register(m)
	}
}
