package redisdsync

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/dsync"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/redis"
	redislib "github.com/go-redis/redis/v8"
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
	AppCtx       *bootstrap.ApplicationContext
	RedisFactory redis.ClientFactory        `optional:"true"`
	RedisClients []redislib.UniversalClient `group:"dsync"`
}

func provideSyncManager(di syncDI) (dsync.SyncManager, error) {
	var clients []redislib.UniversalClient
	switch {
	case len(di.RedisClients) != 0:
		clients = append(clients, di.RedisClients...)
	case di.RedisFactory != nil:
		client, e := di.RedisFactory.New(di.AppCtx, func(cOpt *redis.ClientOption) {
			cOpt.DbIndex = 1
		})
		if e != nil {
			return nil, dsync.ErrSyncManagerStopped.WithMessage("unable to initialize Redis SyncManager").WithCause(e)
		}
		clients = []redislib.UniversalClient{client}
	default:
		return nil, fmt.Errorf(`redis.ClientFactory or []go-redis/redis/v8.UniversalClient with FX group '%s' are required for 'redisdsync' package`, dsync.FxGroup)
	}

	return NewRedisSyncManager(di.AppCtx, func(opt *RedisSyncOption) {
		opt.Clients = clients
	}), nil
}
