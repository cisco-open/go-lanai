package security

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: 1,
	Provides: []fx.Option{fx.Provide(NewBasicAuth)},
	Invokes: []fx.Option{fx.Invoke(setup)},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}
