package data

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("Data")

var Module = &bootstrap.Module{
	Name: "cockroach",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(data.NewGorm),
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
//func initialize(lc fx.Lifecycle) {
//	lc.Append(fx.Hook{
//		OnStart: func(ctx context.Context) error {
//
//		},
//		OnStop:  func(ctx context.Context) error {
//
//		},
//	})
//}



