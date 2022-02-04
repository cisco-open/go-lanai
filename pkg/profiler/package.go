package profiler

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

const (
	RouteGroup      = "debug"
	PathPrefixPProf = "pprof"
)

var Module = &bootstrap.Module{
	Precedence: bootstrap.DebugPrecedence,
	Options: []fx.Option{
		fx.Invoke(initialize),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

type initDI struct {
	fx.In
	Lifecycle    fx.Lifecycle
	WebRegistrar *web.Registrar `optional:"true"`
}

func initialize(di initDI) {
	if di.WebRegistrar == nil {
		return
	}
	di.WebRegistrar.MustRegister(&PProfController{})
}

