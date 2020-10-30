package bootstrap

import (
	"context"
	"fmt"
	"go.uber.org/fx"
)

var applicationContext = NewContext()

var DefaultModule = &Module{
	Precedence: HighestPrecedence,
	PriorityOptions: []fx.Option{
		fx.Supply(applicationContext),
		fx.Invoke(bootstrap),
	},
}

func init() {
	Register(DefaultModule)
}

func bootstrap(lc fx.Lifecycle) {
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
