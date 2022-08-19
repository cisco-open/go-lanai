package opensearch

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
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
	},
}

func Use() {
	bootstrap.Register(Module)
}
