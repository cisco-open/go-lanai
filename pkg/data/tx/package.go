package tx

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var logger = log.New("DB.Tx")

var Module = &bootstrap.Module{
	Name: "DB Tx",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(provideGormTxManager),
		fx.Invoke(setGlobalTxManager),
	},
}

type txManagerOut struct {
	fx.Out
	Tx TxManager
	GormTx GormTxManager
}
func provideGormTxManager(db *gorm.DB) txManagerOut {
	m := newGormTxManager(db)
	return txManagerOut{
		Tx: m,
		GormTx: m,
	}
}
