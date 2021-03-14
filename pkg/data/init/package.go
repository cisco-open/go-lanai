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
	Name: "DB",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(data.NewGorm),
		web.FxErrorTranslatorProviders(
			provideDataErrorTranslator,
			provideGormErrorTranslator,
			providePqErrorTranslator,
		),
	},
}

func init() {
	bootstrap.Register(Module)
	bootstrap.Register(tx.Module)
}

func Use() {

}

/**************************
	Provider
***************************/

/**************************
	Initialize
***************************/




