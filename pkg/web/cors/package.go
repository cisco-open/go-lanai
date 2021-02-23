package cors

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "CORS",
	Precedence: bootstrap.FrameworkModulePrecedence + 1001,
	PriorityOptions: []fx.Option{
		fx.Provide(newCustomizer, BindCorsProperties),
	},
}

func init() {
	bootstrap.Register(Module)
}
