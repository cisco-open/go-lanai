package cockroach

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
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
	bootstrap.Register(tlsconfig.Module)
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

func pqErrorTranslatorProvider() fx.Annotated {
	return fx.Annotated{
		Group:  data.GormConfigurerGroup,
		Target: NewPqErrorTranslator,
	}
}

/**************************
	Initialize
***************************/
