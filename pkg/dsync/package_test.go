package dsync_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/dsync"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/consultest"
    "fmt"
    "github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Tests
 *************************/

type TestModuleDI struct {
	fx.In
	Manager dsync.SyncManager
}

func TestModuleInit(t *testing.T) {
	di := TestModuleDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		consultest.WithHttpPlayback(t,
			//consultest.HttpRecordingMode(),
		),
		apptest.WithModules(dsync.Module),
		apptest.WithFxOptions(),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestGettingLock(&di), "TestGettingLock"),
		test.GomegaSubTest(SubTestLeadershipLock(&di), "TestLeadershipLock"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestGettingLock(di *TestModuleDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const (
			key1 = "test-lock"
			key2 = "test-alt-lock"
		)
		var e error
		var lock dsync.Lock
        // get a lock
		e = recoverable(func() {
		    lock = dsync.LockWithKey(key1)
        })
		g.Expect(e).To(Succeed(), "getting 1st lock should not fail")
		g.Expect(lock).ToNot(BeNil(), "lock should not be nil")

        // get another lock
        e = recoverable(func() {
            lock = dsync.LockWithKey(key2)
        })
        g.Expect(e).To(Succeed(), "getting 2nd lock should not fail")
        g.Expect(lock).ToNot(BeNil(), "lock should not be nil")

        // get same lock
        var another dsync.Lock
        e = recoverable(func() {
            another = dsync.LockWithKey(key2)
        })
        g.Expect(e).To(Succeed(), "re-getting lock should not fail")
        g.Expect(another).To(Equal(lock), "re-getting lock should be same instance")
	}
}

func SubTestLeadershipLock(di *TestModuleDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		lock := dsync.LeadershipLock()
		g.Expect(lock).ToNot(BeNil(), "leadership lock should not be nil")
		e := lock.TryLock(ctx)
		defer func() {_ = lock.Release()}()
		g.Expect(e).To(Succeed(), "leadership lock should not return error when acquired")
	}
}

/*************************
	Helpers
 *************************/

func recoverable(fn func()) (err error) {
    defer fn()
    defer func() {
        if v := recover(); v != nil {
            err = fmt.Errorf("%v", v)
        }
    }()
    return nil
}