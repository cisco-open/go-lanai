package discovery

import (
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/test"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "testing"
    "time"
)

/*************************
	Tests
 *************************/

func TestServiceCache(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestSetWithTTL(), "TestSetWithTTL"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestSetWithTTL() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const name = `testservice`
		var ttl = 250 * time.Millisecond
		cache := newSimpleServiceCache()
		svc := &Service{
			Name: name,
			Time: time.Now(),
		}
		cache.SetWithTTL(name, svc, ttl)
        AssertCacheEntry(g, cache, name, true)
        time.Sleep(ttl)
        AssertCacheEntry(g, cache, name, false)
	}
}

/*************************
	Helpers
 *************************/

func AssertCacheEntry(g *gomega.WithT, cache ServiceCache, name string, expected bool) {
	if expected {
		g.Expect(cache.Has(name)).To(BeTrue(), "cache.Has() should be correct")
		g.Expect(cache.Get(name)).ToNot(BeNil(), "cache.Get() should be correct")
		g.Expect(cache.Entries()).To(HaveKeyWithValue(name, Not(BeNil())), "cache.Entries{} should be correct")
	} else {
        g.Expect(cache.Has(name)).To(BeFalse(), "cache.Has() should be correct")
        g.Expect(cache.Get(name)).To(BeNil(), "cache.Get() should be correct")
        g.Expect(cache.Entries()).ToNot(HaveKeyWithValue(name, Not(BeNil())), "cache.Entries{} should be correct")

    }
}
