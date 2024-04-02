// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
	ErrLockUnavailable      = newError("lock is held by another session")
	ErrUnlockFailed         = newError("failed to release lock")
	ErrSessionUnavailable   = newError("session is not available")
	ErrSyncManagerStopped   = newError("sync manager stopped")
	ErrFailedInitialization = newError("sync manager failed to start")
)

// SyncManager manage distributed locks across the application.
type SyncManager interface {
	// Lock returns a distributed lock with given key. If the Lock already exists with same key,
	// the options are ignored and the same Lock is returned.
	//
	// The returned Lock is goroutines-safe, but locking/releasing same lock from different goroutine may cause
	// complicated scenarios. It's application's responsibility to coordinate such concurrent usage.
	Lock(key string, opts ...LockOptions) (Lock, error)
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
// Once lock acquisition is started (Lock.Lock or Lock.TryLock), regardless the result, the Lock would keep trying
// to acquire/re-acquire the lock until Lock.Release is manually invoked, because the lock might be revoked by operator
// or external infra service.
//
// Long-running goroutine should monitor Lost channel after the lock is acquired.
// When Lost channel is signalled, there is no need to re-invoke Lock.Lock or Lock.TryLock, since internal loop would try
// to re-acquire lock. However, any existing tasks relying on this lock should be stopped because there is no guarantee
// that the lock will be re-acquired
type Lock interface {
	// Key the unique identifier of the lock
	Key() string

	// Lock attempts to acquire the lock and blocks until lock is acquired or context is cancelled/timed out.
	// Invoking Lock after lock is acquired (or re-acquired after some error) returns immediately.
	//
	// A cancellable context.Context can be used to abort the current attempt, but it won't stop the lock to keep
	// trying in the background.
	//
	// It is NOT safe to assume that the lock is guaranteed to be held until Release(). The lock might be lost
	// due to session invalidation, communication errors, operator intervention, etc.
	//
	// Lost() returns a channel that is closed if our lock is lost or an error occurred.
	// By default, dsync implementations prefer liveness over safety and an application must be able to handle
	// the lock being lost.
	//
	// Important: A paring call of Release() is always required regardless the result.
	Lock(ctx context.Context) error

	// TryLock differs from Lock in following ways:
	// - TryLock stop blocking when lock is held by other instance/session
	// - TryLock stop blocking when unrecoverable error happens during lock acquisition
	// Note: TryLock may temporarily block when connectivity to external infra service is not available
	TryLock(ctx context.Context) error

	// Release stops the attempt to acquire the lock and releases the lock if already held
	// Release must be used everytime after Lock or TryLock is called, unless the application is intended
	// to hold the lock indefinitely.
	//
	// Invoking Release multiple time takes no effect.
	//
	// Note: Lost channel would stop signalling after Release, until Lock or TryLock is called again.
	Release() error

	// Lost channel signals long-running goroutine when lock is lost (due to network error, operator intervention,
	// manual Release() call from other goroutine, etc).
	//
	// When Lost channel is signalled, there is no need to re-invoke Lock.Lock or Lock.TryLock for lock re-acquisition
	// unless it's caused by manual Release() call, but all relying-tasks should pause.
	Lost() <-chan struct{}
}

/*********************
	Common Impl
 *********************/

// LockWithKey returns a distributed Lock with given key. If the Lock already exists with same key,
// the options are ignored and the same Lock is returned.
//
// The returned Lock is goroutines-safe, but locking/releasing same lock from different goroutine may cause
// complicated scenarios. It's application's responsibility to coordinate such concurrent usage.
//
// This function panic if internal SyncManager is not initialized yet or key is not provided.
func LockWithKey(key string, opts ...LockOptions) Lock {
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
