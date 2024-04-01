package consuldsync

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/pkg/dsync"
	"github.com/cisco-open/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("DSync")

var Module = &bootstrap.Module{
	Name:       "distributed",
	Precedence: bootstrap.DistributedLockPrecedence,
	Options: []fx.Option{
		fx.Provide(provideSyncManager),
	},
	Modules: []*bootstrap.Module{dsync.Module},
}

func Use() {
	bootstrap.Register(Module)
}

/**************************
	Provider
***************************/

type syncDI struct {
	fx.In
	AppCtx          *bootstrap.ApplicationContext
	Conn            *consul.Connection  `optional:"true"`
	TestSyncManager []dsync.SyncManager `group:"test"`
}

func provideSyncManager(di syncDI) (dsync.SyncManager, error) {
	if len(di.TestSyncManager) != 0 {
		return di.TestSyncManager[0], nil
	}
	if di.Conn == nil {
		return nil, fmt.Errorf("*consul.Connection is required for 'dsync' package")
	}
	return NewConsulLockManager(di.AppCtx, di.Conn), nil
}

/**************************
	Initialize
***************************/
