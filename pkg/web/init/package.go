package init

import (
	"context"
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/cors"
	"embed"
	"go.uber.org/fx"
)

//go:embed defaults-web.yml
var defaultConfigFS embed.FS

var Module = &bootstrap.Module{
	Name: "web",
	Precedence: web.MinWebPrecedence,
	PriorityOptions: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(
			web.BindServerProperties,
			web.NewEngine,
			web.NewRegistrar),
		fx.Invoke(setup),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
	bootstrap.Register(cors.Module)
}

/**************************
	Provide dependencies
***************************/

/**************************
	Setup
***************************/
type initDI struct {
	fx.In
	Registrar        *web.Registrar
	Properties       web.ServerProperties
	Controllers      []web.Controller      `group:"controllers"`
	Customizers      []web.Customizer      `group:"customizers"`
	ErrorTranslators []web.ErrorTranslator `group:"error_translators"`
}

func setup(lc fx.Lifecycle, di initDI) {
	_ = di.Registrar.Register(web.NewLoggingCustomizer(di.Properties))
	_ = di.Registrar.Register(web.NewRecoveryCustomizer())
	_ = di.Registrar.Register(web.NewGinErrorHandlingCustomizer())

	_ = di.Registrar.Register(di.Controllers)
	_ = di.Registrar.Register(di.Customizers)
	_ = di.Registrar.Register(di.ErrorTranslators)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			return di.Registrar.Run(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return di.Registrar.Stop(ctx)
		},
	})
}
