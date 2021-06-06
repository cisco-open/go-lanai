package bootstrap

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("Bootstrap")

func DefaultModule(cliCtx *CliExecContext, app *App) *Module {
	return &Module{
		Precedence: HighestPrecedence,
		PriorityOptions: []fx.Option{
			fx.Logger(newFxPrinter(logger, app)),
			fx.Supply(cliCtx),
			fx.Supply(app),
			fx.Provide(provideApplicationContext),
			fx.Invoke(bootstrap),
		},
	}
}

func provideApplicationContext(app *App, config ApplicationConfig) *ApplicationContext {
	app.ctx.config = config
	return app.ctx
}

func bootstrap(lc fx.Lifecycle, ac *ApplicationContext) {
	logProperties := &log.Properties{}
	err := ac.config.Bind(logProperties, "log")
	if err == nil {
		err = log.UpdateLoggingConfiguration(logProperties)
	}
	if err != nil {
		logger.Error( "Error updating logging configuration", "error", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.WithContext(ac).Info("On Application Start")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
