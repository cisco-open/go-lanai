package bootstrap

import (
	"context"
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig"
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

type provideApplicationContextParam struct {
	fx.In
	Config *appconfig.Config `name:"application_config"`
}
func provideApplicationContext(p provideApplicationContextParam) *ApplicationContext {
	applicationContext.updateConfig(p.Config)
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
