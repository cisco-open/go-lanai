package bootstrap

import (
	"context"
	"fmt"
	"go.uber.org/fx"
)

var bootstrapContext = NewContext()

var DefaultModule = &Module{
	Precedence: HighestPrecedence,
	Provides: []fx.Option{fx.Supply(bootstrapContext)},
	Invokes: []fx.Option{fx.Invoke(bootstrap)},
}

func init() {
	Register(DefaultModule)
}

// no need to use maker func, this package should be always included in main()


func bootstrap(lc fx.Lifecycle) {
	fmt.Println("[bootstrap] - bootstrap")
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("[bootstrap] - OnStart")
			ctx.(*Context).PutValue("key", "value")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}



