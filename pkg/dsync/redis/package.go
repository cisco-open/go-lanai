package redisdsync

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/dsync"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"go.uber.org/fx"
)

var logger = log.New("DSync")

var Module = &bootstrap.Module{
	Name:       "distributed",
	Precedence: bootstrap.DistributedLockPrecedence,
	Options: []fx.Option{
		//appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(provideSyncManager),
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
	Redis  redis.ClientFactory `optional:"true"`
	//TestSyncManager []dsync.SyncManager `group:"test"`
}

func provideSyncManager(di syncDI) (dsync.SyncManager, error) {
	//if len(di.TestSyncManager) != 0 {
	//	return di.TestSyncManager[0], nil
	//}
	if di.Redis == nil {
		return nil, fmt.Errorf("redis.ClientFactory is required for 'redisdsync' package")
	}
	return NewRedisSyncManager(di.AppCtx, di.Redis), nil
}
