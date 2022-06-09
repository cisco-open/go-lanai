package dsyncmock

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSyncManager(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestNoopSyncManager(), "TestNoopSyncManager"),
	)
}

func SubTestNoopSyncManager() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		out := ProvideNoopSyncManager()
		manager := out.TestSyncManager
		l, e := manager.Lock("test-key")
		g.Expect(e).To(Succeed())
		g.Expect(l.Key()).To(BeEquivalentTo("test-key"))
		g.Expect(l.Lock(ctx)).To(Succeed())
		g.Expect(l.TryLock(ctx)).To(Succeed())
		g.Expect(l.Release()).To(Succeed())
		g.Expect(l.Lost()).To(HaveLen(0))
	}
}
