package dsync

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("DSync")

var Module = &bootstrap.Module{
	Name:       "distributed",
	Precedence: bootstrap.DistributedLockPrecedence,
	Options: []fx.Option{
		fx.Provide(provideSyncManager),
		fx.Invoke(setup),
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

func setup(lc fx.Lifecycle, manager SyncManager) {
	syncLc, ok := manager.(SyncManagerLifecycle)
	if !ok {
		return
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if e := syncLc.Start(ctx); e != nil {
				return e
			}
			// TODO start leader election lock
			return nil
		},
		OnStop:  syncLc.Stop,
	})
}
