package webtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var mockedWebModule = &bootstrap.Module{
	Name: "web",
	Precedence: web.MinWebPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(
			web.BindServerProperties,
			web.NewEngine,
			web.NewRegistrar),
		fx.Invoke(initialize),
	},
}


type initDI struct {
	fx.In
	Registrar        *web.Registrar
	Properties       web.ServerProperties
	Controllers      []web.Controller      `group:"controllers"`
	Customizers      []web.Customizer      `group:"customizers"`
	ErrorTranslators []web.ErrorTranslator `group:"error_translators"`
}

func initialize(lc fx.Lifecycle, di initDI) {
	di.Registrar.MustRegister(web.NewLoggingCustomizer(di.Properties))
	di.Registrar.MustRegister(web.NewRecoveryCustomizer())
	di.Registrar.MustRegister(web.NewGinErrorHandlingCustomizer())

	di.Registrar.MustRegister(di.Controllers)
	di.Registrar.MustRegister(di.Customizers)
	di.Registrar.MustRegister(di.ErrorTranslators)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) (err error) {
			if err = di.Registrar.Initialize(ctx); err != nil {
				return
			}
			defer func(ctx context.Context) {
				_ = di.Registrar.Cleanup(ctx)
			}(ctx)
			return nil
		},
	})
}
