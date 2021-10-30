package cockroach

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

//goland:noinspection GoUnusedGlobalVariable
var logger = log.New("CockroachDB")

var Module = &bootstrap.Module{
	Name:       "cockroach",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(NewGormDialetor,
			BindCockroachProperties,
			NewGormDbCreator,
			pqErrorTranslatorProvider(),
		),
		//fx.Invoke(initialize),
	},
}

func Use() {
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

func pqErrorTranslatorProvider() fx.Annotated {
	return fx.Annotated{
		Group: data.GormConfigurerGroup,
		Target: func() data.ErrorTranslator {
			return PqErrorTranslator{}
		},
	}
}

/**************************
	Initialize
***************************/
func initialize(lc fx.Lifecycle) {
	//lc.Append(fx.Hook{
	//	OnStart: func(ctx context.Context) error {
	//
	//	},
	//	OnStop:  func(ctx context.Context) error {
	//
	//	},
	//})
}
