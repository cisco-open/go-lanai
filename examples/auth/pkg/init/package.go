package serviceinit

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/examples/auth-service/pkg/controller"
	"cto-github.cisco.com/NFV-BU/go-lanai/examples/auth-service/pkg/service"
	actuator "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/init"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	consul "cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul/init"
	discoveryinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	tracing "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing/init"
	vault "cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault/init"
	web "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/init"
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
