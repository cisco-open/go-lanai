package redisdsync

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/dsync"
	redislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"sync"
	"time"
)

type RedisSyncOptions func(opt *RedisSyncOption)
type RedisSyncOption struct {
	// Clients are go-redis/v8 clients.
	// Each client should be able to connect to an independent Redis master/cluster/sentinel-master to form quorum
	Clients []redislib.UniversalClient
	// TTL see RedisLockOption.AutoExpiry
	TTL time.Duration
	// RetryDelay see RedisLockOption.RetryDelay
	RetryDelay time.Duration
	// TimeoutFactor see RedisLockOption.TimeoutFactor
	TimeoutFactor float64
}

func NewRedisSyncManager(appCtx *bootstrap.ApplicationContext, opts ...RedisSyncOptions) *RedisSyncManager {
	opt := RedisSyncOption{
		TTL:           10 * time.Second,
		RetryDelay:    1 * time.Second,
		TimeoutFactor: 0.05,
	}
	for _, fn := range opts {
		fn(&opt)
	}

	pools := make([]redsyncredis.Pool, len(opt.Clients))
	for i := range opt.Clients {
		pools[i] = goredis.NewPool(opt.Clients[i])
	}

	return &RedisSyncManager{
		appCtx:  appCtx,
		options: opt,
		syncer:  redsync.New(pools...),
		locks:   make(map[string]*RedisLock),
	}
}

type RedisSyncManager struct {
	appCtx   *bootstrap.ApplicationContext
	options  RedisSyncOption
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

	m.mtx.Lock()
	defer m.mtx.Unlock()

	if lock, ok := m.locks[key]; ok {
		return lock, nil
	}

	m.locks[key] = newRedisLock(m.syncer, func(opt *RedisLockOption) {
		opt.Context = m.appCtx
		opt.Name = key
		opt.AutoExpiry = m.options.TTL
		opt.RetryDelay = m.options.RetryDelay
		opt.TimeoutFactor = m.options.TimeoutFactor
	})
	return m.locks[key], nil
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
