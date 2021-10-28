package data

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/repo"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqcrypt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
	"reflect"
)

//var logger = log.New("Data")

var Module = &bootstrap.Module{
	Name:       "DB",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(data.NewGorm),
		web.FxErrorTranslatorProviders(
			provideDriverErrorTranslator,
			webErrTranslatorProvider(data.NewDataErrorTranslator),
			webErrTranslatorProvider(data.NewGormErrorTranslator),
		),
	},
}

func Use() {
	bootstrap.Register(Module)
	bootstrap.Register(data.Module)
	bootstrap.Register(tx.Module)
	bootstrap.Register(repo.Module)
	bootstrap.Register(pqcrypt.Module)
}

/**************************
	Provider
***************************/

func webErrTranslatorProvider(provider interface{}) func() web.ErrorTranslator {
	return func() web.ErrorTranslator {
		fnv := reflect.ValueOf(provider)
		ret := fnv.Call(nil)
		return ret[0].Interface().(web.ErrorTranslator)
	}
}

type detDI struct {
	fx.In
	Translators []data.ErrorTranslator `group:"data-driver"`
}

func provideDriverErrorTranslator(di detDI) web.ErrorTranslator {
	return data.NewDriverErrorTranslator(di.Translators...)
}

/**************************
	Initialize
***************************/
