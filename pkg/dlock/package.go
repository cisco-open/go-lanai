package dlock

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("DLock")

var Module = &bootstrap.Module{
	Name:       "distributed",
	Precedence: bootstrap.DistributedLockPrecedence,
	Options: []fx.Option{
		fx.Provide(provideSyncManager),
		//fx.Invoke(initialize),
	},
}

func Use() {
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

type syncDI struct {
	fx.In
	AppCtx *bootstrap.ApplicationContext
	Conn   *consul.Connection
}

func provideSyncManager(di syncDI) SyncManager {
	return newConsulLockManager(di.AppCtx, di.Conn)
}

/**************************
	Initialize
***************************/
