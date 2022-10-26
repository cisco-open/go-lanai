package opensearch

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	"go.uber.org/fx"
)

var logger = log.New("Search")

var Module = &bootstrap.Module{
	Precedence: bootstrap.OpenSearchPrecedence,
	Options: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(BindOpenSearchProperties),
		fx.Provide(NewConfig),
		fx.Provide(NewClient),
		fx.Invoke(registerHealth),
	},
}

func Use() {
	bootstrap.Register(Module)
	bootstrap.Register(tlsconfig.Module)
}

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	OpenClient      OpenClient       `optional:"true"`
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil || di.OpenClient == nil {
		return
	}
	di.HealthRegistrar.MustRegister(NewHealthIndicator(di.OpenClient))
}
