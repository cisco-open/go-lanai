package serviceinit

import (
	"github.com/cisco-open/go-lanai/examples/auth-service/pkg/controller"
	"github.com/cisco-open/go-lanai/examples/auth-service/pkg/service"
	actuator "github.com/cisco-open/go-lanai/pkg/actuator/init"
	appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	consul "github.com/cisco-open/go-lanai/pkg/consul/init"
	discoveryinit "github.com/cisco-open/go-lanai/pkg/discovery/init"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/pkg/security/config/authserver"
	"github.com/cisco-open/go-lanai/pkg/security/config/resserver"
	"github.com/cisco-open/go-lanai/pkg/security/idp/passwdidp"
	tracing "github.com/cisco-open/go-lanai/pkg/tracing/init"
	vault "github.com/cisco-open/go-lanai/pkg/vault/init"
	web "github.com/cisco-open/go-lanai/pkg/web/init"
	"go.uber.org/fx"
)

func Use() {
	// basic modules
	appconfig.Use()
	consul.Use()
	vault.Use()
	web.Use()
	redis.Use()
	actuator.Use()
	tracing.Use()
	discoveryinit.Use()

	authserver.Use()
	resserver.Use()
	passwdidp.Use()
	controller.Use()
	service.Use()
	bootstrap.AddOptions(
		fx.Provide(newAuthServerConfigurer),
		fx.Provide(newResServerConfigurer),
		fx.Invoke(configureSecurity),
	)
}
