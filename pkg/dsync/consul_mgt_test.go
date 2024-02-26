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

package dsync_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/pkg/dsync"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/consultest"
	"github.com/cisco-open/go-lanai/test/ittest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

/*************************
	Tests
 *************************/

type TestConsulDsyncDI struct {
	fx.In
	ittest.RecorderDI
	AppCtx *bootstrap.ApplicationContext
	Consul *consul.Connection
}

func TestConsulDSyncManager(t *testing.T) {
	di := TestConsulDsyncDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		consultest.WithHttpPlayback(t,
			//consultest.HttpRecordingMode(),
			// - Too many concurrent operations, ordering is different every time.
			// - Latency is also required because consul lock is heavily rely on blocking HTTP transactions
			consultest.MoreHTTPVCROptions(ittest.DisableHttpRecordOrdering(), ittest.ApplyHttpLatency()),
		),
		apptest.WithTimeout(2*time.Minute),
		apptest.WithFxOptions(),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestConsulTryLock(&di), "TestTryLock"),
		test.GomegaSubTest(SubTestConsulLockAndRelease(&di), "TestLockAndRelease"),
		test.GomegaSubTest(SubTestConsulSessionRecovery(&di), "TestSessionRecovery"),
		test.GomegaSubTest(SubTestConsulInvalidSession(&di), "TestInvalidSession"),
		test.GomegaSubTest(SubTestConsulCancelledContext(&di), "TestCancelledContext"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestConsulTryLock(di *TestConsulDsyncDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const lockKey = "try-lock-test"
		mgts := NewConsulManagers(di, g)
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

func SubTestConsulLockAndRelease(di *TestConsulDsyncDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const lockKey = "lock-test"
		mgts := NewConsulManagers(di, g, func(opt *dsync.ConsulSessionOption) {
			// minimum TTL and lock delay (1sec) for faster test
			opt.TTL = 10 * time.Second
			opt.LockDelay = 1 * time.Second
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

func SubTestConsulSessionRecovery(di *TestConsulDsyncDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const lockKey = "session-renew-test"
		const sessionName = `test-session`
		var ttl = 10 * time.Second
		mgts := NewConsulManagers(di, g, func(opt *dsync.ConsulSessionOption) {
			opt.Name = sessionName
			// minimum TTL and lock delay (1sec) for faster test. Lower retry rate to avoid too many retries.
			opt.TTL = ttl
			opt.LockDelay = 1 * time.Second
			opt.RetryDelay = 3 * time.Second
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
		e = lock1.TryLock(timeoutCtx)
		g.Expect(e).To(Succeed(), "TryLock should not fail when lock is acquirable")

		// remove session
		go RemoveSession(g, di.Consul, sessionName)

		// Wait until revoked, validate Lost() channel works properly
		timeoutCtx, cancelFn = context.WithTimeout(ctx, ttl)
		defer cancelFn()
		select {
		case <-lock1.Lost():
			e = lock1.TryLock(timeoutCtx)
			g.Expect(e).To(HaveOccurred(), "TryLock should fail when session is revoked")
		case <-timeoutCtx.Done():
			t.Errorf("expect signal of lost lock after session revocation, got nothing")
		}

		// Try to re-acquire after session is recovered
		// Note: Session is usually recovered within half of the TTL, see consul's api.Session.RenewPeriodic().
		timeoutCtx, cancelFn = context.WithTimeout(ctx, 6 * time.Second)
		defer cancelFn()
		e = lock1.Lock(timeoutCtx)
		g.Expect(e).To(Succeed(), "Lock should be eventually acquirable after session is recovered")
	}
}

func SubTestConsulInvalidSession(di *TestConsulDsyncDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const lockKey = "invalid-session-test"
		mgts := NewConsulManagers(di, g, func(opt *dsync.ConsulSessionOption) {
			// minimum TTL is 10s. Creating session will time out if TTL is invalid
			opt.TTL = 1 * time.Second
		})
		mgts.Start(ctx, g)
		defer mgts.Stop(ctx, g)

		var timeout = 100 * time.Millisecond
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
		g.Expect(e).To(HaveOccurred(), "Lock should fail when session is not valid")
	}
}

func SubTestConsulCancelledContext(di *TestConsulDsyncDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const lockKey = "cancelled-context-test"
		mgts := NewConsulManagers(di, g)
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
		time.AfterFunc(100 * time.Millisecond, cancelFn)
		time.AfterFunc(200 * time.Millisecond, stopFn1)
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

func RemoveSession(g *WithT, consulConn *consul.Connection, sessionName string) {
	entries, _, e := consulConn.Client().Session().List(nil)
	g.Expect(e).To(Succeed(), "listing current sessions should not fail")
	var sid string
	for i := range entries {
		if entries[i].Name == sessionName {
			sid = entries[i].ID
			break
		}
	}
	g.Expect(sid).ToNot(BeEmpty(), "session with name [%s] should exist", sessionName)
	_, e = consulConn.Client().Session().Destroy(sid, nil)
	g.Expect(e).To(Succeed(), "deleting session [%s](%s) should not fail", sessionName, sid)
}

// TestConsulManagers to mimic distributed environment, we always needs multiple managers
type TestConsulManagers struct {
	Main      *dsync.ConsulSyncManager
	Secondary *dsync.ConsulSyncManager
}

func NewConsulManagers(di *TestConsulDsyncDI, g *gomega.WithT, opts ...dsync.ConsulSessionOptions) TestConsulManagers {
	ret := TestConsulManagers{
		Main:      dsync.NewConsulLockManager(di.AppCtx, di.Consul, opts...),
		Secondary: dsync.NewConsulLockManager(di.AppCtx, di.Consul, opts...),
	}
	g.Expect(ret.Main).ToNot(BeNil(), "major consul sync manager should not be nil")
	g.Expect(ret.Secondary).ToNot(BeNil(), "minor consul sync manager should not be nil")
	return ret
}

func (m TestConsulManagers) Start(ctx context.Context, g *gomega.WithT) {
	e := m.Main.Start(ctx)
	g.Expect(e).To(Succeed(), "starting major manager should not fail")
	e = m.Secondary.Start(ctx)
	g.Expect(e).To(Succeed(), "starting minor manager should not fail")
}

func (m TestConsulManagers) Stop(ctx context.Context, g *gomega.WithT) {
	e := m.Main.Stop(ctx)
	g.Expect(e).To(Succeed(), "stopping major manager should not fail")
	e = m.Secondary.Stop(ctx)
	g.Expect(e).To(Succeed(), "stopping minor manager should not fail")
}
