package bootstrap

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var applicationContext = NewContext()
var logger = log.New("bootstrap")

var DefaultModule = &Module{
	Precedence: HighestPrecedence,
	PriorityOptions: []fx.Option{
		fx.Provide(provideApplicationContext),
		fx.Invoke(bootstrap),
	},
}

func init() {
	Register(DefaultModule)
}

func provideApplicationContext(config *appconfig.ApplicationConfig) *ApplicationContext {
	applicationContext.updateConfig(config)
	return applicationContext
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
			logger.Info("On Application Start")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
