package bootstrap

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"fmt"
	"go.uber.org/fx"
)

var applicationContext = NewContext()

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
	fmt.Println("[bootstrap] - bootstrap")

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ac := ctx.(*ApplicationContext)

			fmt.Println("[bootstrap] - On Application Start")
			ac.dumpConfigurations()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
