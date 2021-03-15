package init

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/cors"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "web",
	Precedence: web.MinWebPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(web.BindServerProperties,
			web.NewEngine,
			web.NewRegistrar),
		fx.Invoke(setup),
	},
}

func init() {
	bootstrap.Register(Module)
	bootstrap.Register(cors.Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

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
	di.Registrar.Register(web.NewLoggingCustomizer(di.Properties))
	di.Registrar.Register(web.NewRecoveryCustomizer())
	di.Registrar.Register(web.NewGinErrorHandlingCustomizer())

	di.Registrar.Register(di.Controllers)
	di.Registrar.Register(di.Customizers)
	di.Registrar.Register(di.ErrorTranslators)

	lc.Append(fx.Hook{
		OnStart: makeMappingRegistrationOnStartHandler(&di),
	})
}

func makeMappingRegistrationOnStartHandler(dep *initDI) bootstrap.LifecycleHandler {
	return func(ctx context.Context) (err error) {
		return dep.Registrar.Run(ctx)
	}
}