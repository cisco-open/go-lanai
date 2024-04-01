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

package redisdsync_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/dsync"
	redisdsync "github.com/cisco-open/go-lanai/pkg/dsync/redis"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/embedded"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

/*************************
	Tests
 *************************/

type TestRedisDsyncDI struct {
	fx.In
	AppCtx *bootstrap.ApplicationContext
	Redis  redis.ClientFactory
}

func TestRedisDSyncManager(t *testing.T) {
	di := TestRedisDsyncDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		embedded.WithRedis(),
		apptest.WithModules(redis.Module),
		//apptest.WithTimeout(2*time.Minute),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestTryLock(&di), "TestTryLock"),
		test.GomegaSubTest(SubTestLockAndRelease(&di), "TestLockAndRelease"),
		test.GomegaSubTest(SubTestLockRecovery(&di, true), "TestRedisDownRecovery"),
		test.GomegaSubTest(SubTestLockRecovery(&di, false), "TestRedisErrorRecovery"),
		test.GomegaSubTest(SubTestCancelledContext(&di), "TestCancelledContext"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestTryLock(di *TestRedisDsyncDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const lockKey = "try-lock-test"
		mgts := NewSyncManagers(di, g)
		mgts.Start(ctx, g)
		defer mgts.Stop(ctx, g)

		var timeout = 1000 * time.Millisecond
		var timeoutCtx context.Context
		var cancelFn context.CancelFunc
		var e error
		var lock1, lock2 dsync.Lock
		var stopFn func()

		// obtain locks
		lock1, stopFn = GetTestLock(g, mgts.Main, lockKey)
		defer stopFn()
		lock2, stopFn = GetTestLock(g, mgts.Secondary, lockKey)
		defer stopFn()

		// main lock - 1st pass
		timeoutCtx, cancelFn = context.WithTimeout(ctx, timeout)
		defer cancelFn()
		e = lock1.TryLock(timeoutCtx)
		g.Expect(e).To(Succeed(), "TryLock should not fail when lock is acquirable")

		// minor lock - try
		timeoutCtx, cancelFn = context.WithTimeout(ctx, timeout)
		defer cancelFn()
		e = lock2.TryLock(timeoutCtx)
		g.Expect(e).To(HaveOccurred(), "TryLock should fail when lock is not acquirable")

		// 2nd pass
		timeoutCtx, cancelFn = context.WithTimeout(ctx, timeout)
		defer cancelFn()
		e = lock1.TryLock(timeoutCtx)
		g.Expect(e).To(Succeed(), "TryLock should not fail when lock is already acquired")
	}
}

func SubTestLockAndRelease(di *TestRedisDsyncDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const lockKey = "lock-test"
		mgts := NewSyncManagers(di, g, func(opt *redisdsync.RedisSyncOption) {
			opt.TTL = 1 * time.Second
			opt.RetryDelay = 10 * time.Millisecond
		})
		mgts.Start(ctx, g)
		defer mgts.Stop(ctx, g)

		var timeout = 200 * time.Millisecond
		var timeoutCtx context.Context
		var cancelFn context.CancelFunc
		var e error
		var lock1, lock2 dsync.Lock
		var stopFn1, stopFn2 func()

		// obtain locks
		lock1, stopFn1 = GetTestLock(g, mgts.Main, lockKey)
		lock2, stopFn2 = GetTestLock(g, mgts.Secondary, lockKey)

		// main lock - acquire
		timeoutCtx, cancelFn = context.WithTimeout(ctx, timeout)
		defer cancelFn()
		e = lock1.Lock(timeoutCtx)
		g.Expect(e).To(Succeed(), "Lock should not fail when lock is acquirable")

		// minor lock - acquire
		timeoutCtx, cancelFn = context.WithTimeout(ctx, timeout)
		defer cancelFn()
		e = lock2.Lock(timeoutCtx)
		g.Expect(e).To(HaveOccurred(), "Lock should timeout when lock is not acquirable")

		// release main lock in another thread, wait until released or timed out
		go stopFn1()
		timeoutCtx, cancelFn = context.WithTimeout(ctx, 5000*time.Millisecond)
		defer cancelFn()
		e = lock2.Lock(timeoutCtx)
		defer stopFn2()
		g.Expect(e).To(Succeed(), "Lock should not fail after lock is released")
	}
}

func SubTestLockRecovery(di *TestRedisDsyncDI, serverDown bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const lockKey = "lock-recovery-test"
		var ttl = 300 * time.Millisecond
		mgts := NewSyncManagers(di, g, func(opt *redisdsync.RedisSyncOption) {
			opt.TTL = ttl
			opt.RetryDelay = 10 * time.Millisecond
			opt.TimeoutFactor = 0.01
		})
		mgts.Start(ctx, g)
		defer mgts.Stop(ctx, g)

		var timeout = 1000 * time.Millisecond
		var timeoutCtx context.Context
		var cancelFn context.CancelFunc
		var e error
		var lock1 dsync.Lock
		var stopFn func()

		// obtain locks
		lock1, stopFn = GetTestLock(g, mgts.Main, lockKey)
		defer stopFn()

		// main lock - acquire
		timeoutCtx, cancelFn = context.WithTimeout(ctx, timeout)
		defer cancelFn()
		e = lock1.Lock(timeoutCtx)
		g.Expect(e).To(Succeed(), "TryLock should not fail when lock is acquirable")

		// cause some interruption
		if serverDown {
			go StopRedisServer(ctx, g)
			defer RestartRedisServer(ctx, g) // just in case
		} else {
			go MockRedisServerError(ctx, g, "oops")
			defer MockRedisServerError(ctx, g, "")
		}


		// Wait until revoked, validate Lost() channel works properly
		timeoutCtx, cancelFn = context.WithTimeout(ctx, ttl*2)
		defer cancelFn()
		select {
		case <-lock1.Lost():
			e = lock1.TryLock(timeoutCtx)
			g.Expect(e).To(HaveOccurred(), "TryLock should fail when redis become unavailable")
		case <-timeoutCtx.Done():
			t.Errorf("expect signal of lost lock after session revocation, got nothing")
		}

		// finish interruption
		if serverDown {
			go RestartRedisServer(ctx, g)
		} else {
			go MockRedisServerError(ctx, g, "")
		}


		// Try to re-acquire after redis is recovered
		timeoutCtx, cancelFn = context.WithTimeout(ctx, 5*time.Second)
		defer cancelFn()
		e = lock1.Lock(timeoutCtx)
		defer stopFn()
		g.Expect(e).To(Succeed(), "Lock should be eventually acquirable after session is recovered")
	}
}

func SubTestCancelledContext(di *TestRedisDsyncDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const lockKey = "cancelled-context-test"
		mgts := NewSyncManagers(di, g)
		mgts.Start(ctx, g)
		defer mgts.Stop(ctx, g)

		var timeout = 1000 * time.Millisecond
		var timeoutCtx context.Context
		var cancelFn context.CancelFunc
		var e error
		var lock1, lock2 dsync.Lock
		var stopFn1, stopFn2 func()

		// obtain locks
		lock1, stopFn1 = GetTestLock(g, mgts.Main, lockKey)
		lock2, stopFn2 = GetTestLock(g, mgts.Secondary, lockKey)

		// main lock - acquire
		timeoutCtx, cancelFn = context.WithTimeout(ctx, timeout)
		defer cancelFn()
		e = lock1.Lock(timeoutCtx)
		g.Expect(e).To(Succeed(), "Lock should not fail when lock is acquirable")

		// release lock1 AFTER context is cancelled,
		timeoutCtx, cancelFn = context.WithTimeout(ctx, timeout)
		time.AfterFunc(100*time.Millisecond, cancelFn)
		time.AfterFunc(200*time.Millisecond, stopFn1)
		e = lock2.Lock(timeoutCtx)
		defer stopFn2()
		g.Expect(e).To(HaveOccurred(), "Lock should fail before the lock become acquirable (cancelled)")
	}
}

/*************************
	Helpers
 *************************/

func GetTestLock(g *WithT, manager dsync.SyncManager, lockName string, opts ...dsync.LockOptions) (dsync.Lock, func()) {
	lock, e := manager.Lock(lockName, opts...)
	g.Expect(e).To(Succeed(), "getting lock should not fail")
	g.Expect(lock).ToNot(BeNil(), "getting lock should not return nil")
	return lock, func() {
		e := lock.Release()
		g.Expect(e).To(Succeed(), "Release should not fail")
	}
}

func StopRedisServer(ctx context.Context, g *WithT) {
	server := embedded.CurrentRedisServer(ctx)
	g.Expect(server).ToNot(BeNil(), "embedded redis server should be available")
	server.Close()
}

func RestartRedisServer(ctx context.Context, g *WithT) {
	server := embedded.CurrentRedisServer(ctx)
	g.Expect(server).ToNot(BeNil(), "embedded redis server should be available")
	server.Lock()
	stopped := server.Server() == nil
	server.Unlock()
	if stopped {
		e := server.Restart()
		g.Expect(e).To(Succeed(), "restart embedded redis should not fail")
	}
}

func MockRedisServerError(ctx context.Context, g *WithT, errMsg string) {
	server := embedded.CurrentRedisServer(ctx)
	g.Expect(server).ToNot(BeNil(), "embedded redis server should be available")
	server.SetError(errMsg)
}

// TestRedisManagers to mimic distributed environment, we always needs multiple managers
type TestRedisManagers struct {
	Main      *redisdsync.RedisSyncManager
	Secondary *redisdsync.RedisSyncManager
}

func NewSyncManagers(di *TestRedisDsyncDI, g *gomega.WithT, opts ...redisdsync.RedisSyncOptions) TestRedisManagers {
	ret := TestRedisManagers{
		Main:      redisdsync.NewRedisSyncManager(di.AppCtx, di.Redis, opts...),
		Secondary: redisdsync.NewRedisSyncManager(di.AppCtx, di.Redis, opts...),
	}
	g.Expect(ret.Main).ToNot(BeNil(), "major consul sync manager should not be nil")
	g.Expect(ret.Secondary).ToNot(BeNil(), "minor consul sync manager should not be nil")
	return ret
}

func (m TestRedisManagers) Start(ctx context.Context, g *gomega.WithT) {
	e := m.Main.Start(ctx)
	g.Expect(e).To(Succeed(), "starting major manager should not fail")
	e = m.Secondary.Start(ctx)
	g.Expect(e).To(Succeed(), "starting minor manager should not fail")
}

func (m TestRedisManagers) Stop(ctx context.Context, g *gomega.WithT) {
	e := m.Main.Stop(ctx)
	g.Expect(e).To(Succeed(), "stopping major manager should not fail")
	e = m.Secondary.Stop(ctx)
	g.Expect(e).To(Succeed(), "stopping minor manager should not fail")
}
