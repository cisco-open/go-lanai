package tx

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

//var logger = log.New("DB.Tx")

var Module = &bootstrap.Module{
	Name:       "DB Tx",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(provideGormTxManager),
		fx.Invoke(setGlobalTxManager),
	},
}

const (
	FxTransactionExecuterOption = "TransactionExecuterOption"
)

type txDI struct {
	fx.In
	UnnamedTx TxManager                   `optional:"true"`
	DB        *gorm.DB                    `optional:"true"`
	Executer  TransactionExecuter         `optional:"true"`
	Options   []TransactionExecuterOption `group:"TransactionExecuterOption"`
}

type txManagerOut struct {
	fx.Out
	Tx     TxManager `name:"tx/TxManager"`
	GormTx GormTxManager
}

func provideGormTxManager(di txDI) txManagerOut {
	// due to limitation of uber/fx, we cannot override provider, which is not good for testing & mocking
	// the workaround is we always use Named Provider as default,
	// then bail the initialization if an Unnamed one is present
	if di.UnnamedTx != nil {
		if override, ok := di.UnnamedTx.(GormTxManager); ok {
			return txManagerOut{Tx: override, GormTx: override}
		} else {
			// we should avoid this path
			return txManagerOut{Tx: di.UnnamedTx, GormTx: gormTxManagerAdapter{TxManager: di.UnnamedTx}}
		}
	}

	if di.DB == nil {
		panic("default GormTxManager requires a *gorm.DB")
	}

	if di.Executer == nil {
		di.Executer = NewDefaultExecuter(di.Options...)
	}
	m := newGormTxManager(di.DB, di.Executer)
	return txManagerOut{
		Tx:     m,
		GormTx: m,
	}
}
