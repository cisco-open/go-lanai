package bootstrap

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var applicationContext = NewApplicationContext()
var logger = log.New("Bootstrap")

var DefaultModule = &Module{
	Precedence: HighestPrecedence,
	PriorityOptions: []fx.Option{
		fx.Logger(&fxPrinter{logger: logger}),
		fx.Provide(provideApplicationContext),
		fx.Invoke(bootstrap),
	},
}

func init() {
	Register(DefaultModule)
}

// EagerGetApplicationContext returns the global ApplicationContext before it becomes available for dependency injection
// Important: packages should typlically get ApplicationContext via fx's dependency injection,
//			  which internal application config are garanteed.
//			  Only packages involved in priority bootstrap (appconfig, consul, vault, etc)
//			  should use this function for logging purpose
// Note: ApplicationContext is made
func EagerGetApplicationContext() *ApplicationContext {
	return applicationContext
}


func provideApplicationContext(config ApplicationConfig) *ApplicationContext {
	applicationContext.config = config
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
			logger.WithContext(ac).Info("On Application Start")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
