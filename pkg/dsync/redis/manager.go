package redisdsync

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/dsync"
	"github.com/cisco-open/go-lanai/pkg/redis"
	redislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"sync"
	"time"
)

type RedisSyncOptions func(opt *RedisSyncOption)
type RedisSyncOption struct {
	Name       string
	TTL        time.Duration
	LockDelay  time.Duration
	RetryDelay time.Duration
	DB         int
}

func NewRedisSyncManager(appCtx *bootstrap.ApplicationContext, factory redis.ClientFactory, opts ...RedisSyncOptions) *RedisSyncManager {
	opt := RedisSyncOption{
		DB: 0,
	}
	for _, fn := range opts {
		fn(&opt)
	}

	return &RedisSyncManager{
		appCtx:  appCtx,
		options: opt,
		factory: factory,
		locks:   make(map[string]*RedisLock),
	}
}

type RedisSyncManager struct {
	initOnce sync.Once
	appCtx   *bootstrap.ApplicationContext
	options  RedisSyncOption
	factory  redis.ClientFactory
	mtx      sync.Mutex
	syncer   *redsync.Redsync
	locks    map[string]*RedisLock
}

func (m *RedisSyncManager) Lock(key string, opts ...dsync.LockOptions) (dsync.Lock, error) {
	if key == "" {
		return nil, fmt.Errorf(`cannot create distributed lock: key is required but missing`)
	}

	opt := dsync.LockOption{
		Valuer: dsync.NewJsonLockValuer(map[string]string{
			"name": fmt.Sprintf("distributed lock - %s", m.appCtx.Name()),
		}),
	}
	for _, fn := range opts {
		fn(&opt)
	}

	if e := m.lazyInit(m.appCtx); e != nil {
		return nil, e
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	if lock, ok := m.locks[key]; ok {
		return lock, nil
	}

	m.locks[key] = newRedisLock(m.syncer, func(opt *RedisLockOption) {
		opt.Context = m.appCtx
		opt.Name = key
	})
	return m.locks[key], nil
}

func (m *RedisSyncManager) lazyInit(ctx context.Context) (err error) {
	m.initOnce.Do(func() {
		client, e := m.factory.New(ctx, func(cOpt *redis.ClientOption) {
			cOpt.DbIndex = m.options.DB
		})
		if e != nil {
			err = dsync.ErrSyncManagerStopped.WithMessage("unable to initialize").WithCause(e)
			return
		}
		pool := goredis.NewPool(redislib.UniversalClient(client))
		m.syncer = redsync.New(pool)
	})
	return
}

func (m *RedisSyncManager) Start(_ context.Context) error {
	return nil
}

func (m *RedisSyncManager) Stop(_ context.Context) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	var failed []string
	for k, lock := range m.locks {
		if e := lock.Release(); e != nil {
			failed = append(failed, k)
		}
	}
	if len(failed) > 0 {
		return dsync.ErrUnlockFailed.WithMessage(`unable to release locks %v`, failed)
	}
	return nil
}
