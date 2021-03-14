package tx

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("DB.Tx")

var Module = &bootstrap.Module{
	Name: "DB Tx",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(newGormTxManager),
		fx.Invoke(setGlobalTxManager),
	},
}
