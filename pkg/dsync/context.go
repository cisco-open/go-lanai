// Package dsync
// Provides distributed synchronization support of microservices and provide common usage patterns
// around distributed lock, such as lock-based service leader election.
package dsync

import (
	"context"
	"encoding/json"
	"fmt"
)

var (
	ErrLockUnavailable    = fmt.Errorf("lock is held by another session")
	ErrSessionUnavailable = fmt.Errorf("session is not available")
	ErrSyncManagerStopped = fmt.Errorf("sync manager stopped")
)

type SyncManager interface {
	// Lock returns a distributed lock for given key.
	// For same key, the same Lock is returned. The returned Lock is goroutines-safe
	// Note: the returned Lock is in idle mode
	Lock(key string, opts...LockOptions) (Lock, error)
}

type SyncManagerLifecycle interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type LockOptions func(opt *LockOption)
type LockOption struct {
	Valuer LockValuer
}

// LockValuer is used to annotate the lock in external infra service.
// It's treated literally and serves as lock's metadata
type LockValuer func() []byte

// Lock distributed mutex lock backed by external infrastructure service such as consul or redis.
// After the lock is acquired (Lock.Lock or Lock.TryLock returns without error), the lock might be revoked by operator or external infra service.
// The Lock would keep trying to acquire/re-acquire the lock until Lock.Release is manually invoked.
//
// Long-running goroutine should monitor Lost channel after the lock is acquired.
// When Lost channel is signalled, there is no need to re-invoke Lock.Lock or Lock.TryLock, since internal loop would try
// to re-acquire lock. However, any existing tasks relying on this lock should be stopped because there is no guarantee
// that the lock will be re-acquired
type Lock interface {
	// Key the unique identifier of the lock
	Key() string

	// Lock blocks until lock is acquired or context is cancelled/timed out.
	// Invoking Lock after lock is acquired (or re-acquired after some error) returns immediately
	Lock(ctx context.Context) error

	// TryLock differs from Lock in following ways:
	// - TryLock stop loop blocking when lock is held by other instance/session
	// - TryLock stop loop blocking when unrecoverable error happens during lock acquisition
	// Note: TryLock may temporarily block when connectivity to external infra service is not available
	TryLock(ctx context.Context) error

	// Release releases the lock. Stop the process from maintaining the active lock.
	// Release must be used after Lock or TryLock is used. Invoking Release multiple time takes no effect
	// Note: Lost channel would stopLoop signalling after Release, until Lock or TryLock is called again
	Release() error

	// Lost channel signals long-running goroutine when lock is lost (due to network error or operator intervention)
	// When Lost channel is signalled, there is no need to re-invoke Lock.Lock or Lock.TryLock for lock re-acquisition,
	// but all relying-tasks should stopLoop.
	Lost() <-chan struct{}
}


/*********************
	Common Impl
 *********************/

// LockWithKey returns a distributed Lock with given key
// this function panic if internal SyncManager is not initialized yet or key is not provided
func LockWithKey(key string, opts...LockOptions) Lock {
	if syncManager == nil {
		panic("SyncManager is not initialized")
	}
	l, e := syncManager.Lock(key, opts...)
	if e != nil {
		panic(e)
	}
	return l
}

// NewJsonLockValuer is the default implementation of LockValuer.
func NewJsonLockValuer(v interface{}) LockValuer {
	return func() []byte {
		data, e := json.Marshal(v)
		if e != nil {
			return []byte(fmt.Sprintf(`"marshalling error: %v"`, e))
		}
		return data
	}
}
