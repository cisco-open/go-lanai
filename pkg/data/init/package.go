package data

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var logger = log.New("Data")

var Module = &bootstrap.Module{
	Name: "cockroach",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(data.NewGorm),
		web.FxErrorTranslatorProviders(
			provideDataErrorTranslator,
			provideGormErrorTranslator,
			providePqErrorTranslator,
		),
		fx.Provide(tx.NewGormTxManager),
		fx.Invoke(tx.SetTxManager),
	},
}

func init() {
	bootstrap.Register(Module)
}

func Use() {

}

/**************************
	Provider
***************************/

/**************************
	Initialize
***************************/




