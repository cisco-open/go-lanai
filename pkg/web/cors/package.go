package cors

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "CORS",
	Precedence: web.MinWebPrecedence + 1,
	PriorityOptions: []fx.Option{
		fx.Provide(BindCorsProperties),
		web.FxCustomizerProviders(newCustomizer),
	},
}

//func init() {
//	bootstrap.Register(Module)
//}

