package actuator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/alive"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/apilist"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/env"
	health "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health/endpoint"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/info"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/loggers"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"embed"
	"go.uber.org/fx"
)

//go:embed defaults-actuator.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name: "actuate-config",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Invoke(info.Register),
		fx.Invoke(health.Register),
		fx.Invoke(env.Register),
		fx.Invoke(alive.Register),
		fx.Invoke(apilist.Register),
		fx.Invoke(loggers.Register),
	},
}

func Use() {
	bootstrap.Register(actuator.Module)
	bootstrap.Register(Module)
}

/**************************
	Initialize
***************************/



