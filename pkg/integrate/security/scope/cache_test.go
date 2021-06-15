package scope

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	test "cto-github.cisco.com/NFV-BU/go-lanai/test/utils"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type cacheCounter struct {
	lc     uint64
	vc uint64
}

func (c *cacheCounter) reset() {
	*c = cacheCounter{}
}

func (c *cacheCounter) countLoad(fn loadFunc) loadFunc {
	return func(ctx context.Context, k cKey) (entryValue, time.Time, error) {
		atomic.AddUint64(&c.lc, 1)
		return fn(ctx, k)
	}
}

func (c *cacheCounter) countValidate(fn validateFunc) validateFunc {
	return func(ctx context.Context, v entryValue) bool {
		atomic.AddUint64(&c.vc, 1)
		return fn(ctx, v)
	}
}

func (c *cacheCounter) loadCount() int {
	return int(atomic.LoadUint64(&c.lc))
}

func (c *cacheCounter) validateCount() int {
	return int(atomic.LoadUint64(&c.vc))
}

/*************************
	Tests
 *************************/
func TestCacheCorrectnessPositiveCases(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestCacheSuccessfulLoad(), "CacheSuccessfulLoad"),
		test.GomegaSubTest(SubTestCacheFailedLoad(), "CacheFailedLoad"),
		test.GomegaSubTest(SubTestCacheOnDifferentKeys(), "CacheOnDifferentKeys"),
		test.GomegaSubTest(SubTestCacheEvict(), "CacheEvict"),
	)
}

//func TestCacheCorrectnessNegativeCases(t *testing.T) {
//	test.RunTest(context.Background(), t,
//		//test.GomegaSubTest(SubTestDefaultDI(bDI, acDI), "SubTestDefaultDI-Pass1"),
//	)
//}

func TestCacheConcurrency(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestCacheConcurrentSoapTest(), "CacheConcurrentSoapTest"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestCacheSuccessfulLoad() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const repeat = 5
		c := newCache()
		loader, expected := staticLoadFunc(100 * time.Millisecond, 60 * time.Second)

		k := cKey{username:   "u1"}
		counter := &cacheCounter{}
		validator := fixedValidateFunc(true)
		auth, _ := testGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			true, BeIdenticalTo(expected), "at first time")

		// try to get while the previous data is still valid
		_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			repeat, true, BeIdenticalTo(expected), "after first load")

		g.Expect(counter.loadCount()).To(Equal(1), "repeated GetOrLoad should only invoke loader once")
		g.Expect(counter.validateCount()).To(Equal(repeat), "validator should be invoked once per repeated GetOrLoad invocation")

		// try to get with previous data invalid
		counter.reset()
		loader, expected = staticLoadFunc(100 * time.Millisecond, 60 * time.Second)
		validator = notValidateFunc(auth)
		_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			repeat, true, And(BeIdenticalTo(expected), Not(BeIdenticalTo(auth))), "after invalidation")

		g.Expect(counter.loadCount()).To(Equal(1), "repeated GetOrLoad should only invoke loader once")
		g.Expect(counter.validateCount()).To(Equal(repeat), "validator should be invoked once per repeated GetOrLoad invocation after invalidation")
	}
}

func SubTestCacheFailedLoad() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const repeat = 5
		c := newCache()
		loader, _ := staticErrLoadFunc(100 * time.Millisecond, 200 * time.Millisecond)

		k := cKey{username:   "u1"}
		counter := &cacheCounter{}
		validator := fixedValidateFunc(true)
		_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			repeat, false, nil, "after first load")

		g.Expect(counter.loadCount()).To(Equal(1), "repeated GetOrLoad should only invoke loader once")
		g.Expect(counter.validateCount()).To(Equal(0), "validator should not be invoked if loader failed")

		// try to get with previous data invalid
		counter.reset()
		time.Sleep(200 * time.Millisecond)
		loader, _ = staticErrLoadFunc(200 * time.Millisecond, 200 * time.Millisecond)
		validator = fixedValidateFunc(false)
		_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			5, false, nil, "after invalidation")

		g.Expect(counter.loadCount()).To(Equal(1), "repeated GetOrLoad should only invoke loader once")
		g.Expect(counter.validateCount()).To(Equal(0), "validator should not be invoked if loader failed")
	}
}

func SubTestCacheOnDifferentKeys() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const keys = 3
		const repeat = 5
		c := newCache()
		validator := fixedValidateFunc(true)

		counters := make([]*cacheCounter, keys)
		results := make([]entryValue, keys)
		for i := 0; i < keys; i++ {
			k := cKey{username:   fmt.Sprintf("u%d", i)}
			counters[i] = &cacheCounter{}
			loader, expected := staticLoadFunc(100 * time.Millisecond, 60 * time.Second)
			results[i], _ = testRepeatedGetOrLoad(ctx, g, c, &k, counters[i].countLoad(loader), counters[i].countValidate(validator),
				repeat, true, BeIdenticalTo(expected), "for " + k.username)

		}

		// assert
		for i := 0; i < keys; i++ {
			for j := i + 1; j < keys; j++ {
				g.Expect(results[i]).To(Not(BeIdenticalTo(results[j])), "different key should yield different value")
			}
			g.Expect(counters[i].loadCount()).To(Equal(1), "repeated GetOrLoad of k%d should only invoke loader once", i)
			g.Expect(counters[i].validateCount()).To(Equal(repeat - 1), "validator of k%d should be invoked once per repeated GetOrLoad invocation", i)
		}
	}
}

func SubTestCacheEvict() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const longExpKeys = 3
		const shortExpKeys = 3
		const repeat = 5
		var exp = 200 * time.Millisecond
		c := newCache(func(opt *cacheOption) {
			opt.Heartbeat = exp / 2
		})
		validator := fixedValidateFunc(true)

		for i := 0; i < shortExpKeys; i++ {
			k := cKey{username:   fmt.Sprintf("su%d", i)}
			loader, expected := staticLoadFunc(100 * time.Millisecond, exp)
			_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, loader, validator,
				repeat, true, BeIdenticalTo(expected), "for " + k.username)
		}

		for i := 0; i < longExpKeys; i++ {
			k := cKey{username:   fmt.Sprintf("lu%d", i)}
			loader, expected := staticLoadFunc(100 * time.Millisecond, 60 * time.Second)
			_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, loader, validator,
				repeat, true, BeIdenticalTo(expected), "for " + k.username)
		}

		time.Sleep(exp)
		g.Expect(len(c.store)).To(Equal(3), "invalidated(expired) entries should be removed")
	}
}

func SubTestCacheConcurrentSoapTest() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {

		// Setup concurrent scenarios
		timeout := 800 * time.Millisecond

		longVLoader, expectedLongV := staticLoadFunc(100 * time.Millisecond, 60 * time.Second)
		shortVLoader, expectedShortV := staticLoadFunc(100 * time.Millisecond, 300 * time.Millisecond)
		unstableLoader, expectedUnstable := staticLoadFunc(100 * time.Millisecond, 60 * time.Second)
		errLoader, _ := staticErrLoadFunc(100 * time.Millisecond, 60 * time.Second)

		stableValidator := fixedValidateFunc(true)
		unstableValidator, ticker := failOccasionallyValidateFunc(2, 200 * time.Millisecond)
		defer ticker.Stop()

		c := newCache(func(opt *cacheOption) {
			opt.Heartbeat = 300 * time.Millisecond
		})
		params := make([]*testCacheParams, 4)
		counters := make([]*cacheCounter, 4)
		params[0], counters[0] = newSuccessCacheParams(c, "long-validity-user", longVLoader, stableValidator, expectedLongV, nil)
		params[1], counters[1] = newSuccessCacheParams(c, "unstable-user", unstableLoader, unstableValidator, expectedUnstable, nil)
		params[2], counters[2] = newFailedCacheParams(c, "non-exist-user", errLoader, stableValidator)
		params[3], counters[3] = newSuccessCacheParams(c, "short-validity-user", shortVLoader, stableValidator, expectedShortV, nil)

		// first, trigger GetOrLoad with special key and short exp period (for evict)
		_, _ = testGetOrLoad(ctx, g, c, &cKey{username:   "to-evicted-user"}, shortVLoader, stableValidator,
			true, BeIdenticalTo(expectedShortV), " for to-evicted-user")

		// Run concurrent test
		count, e := testRepeatedConcurrentGetOrLoad(ctx, g, params, timeout)
		if e != nil {
			t.Errorf("%v", e)
		}

		// Assert invocation count
		g.Expect(count).To(BeNumerically(">", 500), "GetOrLoad should be invoked many times")
		g.Expect(counters[0].loadCount()).To(Equal(1), "repeated GetOrLoad of long validity should only invoke loader once")
		g.Expect(counters[1].loadCount()).To(Equal(3), "repeated GetOrLoad of unstable result should invoke loader 3 times (invalidated twice)")
		g.Expect(counters[2].loadCount()).To(Equal(1), "repeated GetOrLoad of error result should invoke loader once")
		g.Expect(counters[3].loadCount()).To(BeNumerically(">=", 2), "repeated GetOrLoad of short validity should invoke loader more than once")
		g.Expect(counters[3].loadCount()).To(BeNumerically("<=", 10), "repeated GetOrLoad of short validity should not invoke loader too many times ( <= 10)")
		g.Expect(len(c.store)).To(BeNumerically("<=", 4), "invalidated(expired) entries should be removed")
		fmt.Printf("load counts: %d %d %d %d\n", counters[0].loadCount(), counters[1].loadCount(), counters[2].loadCount(), counters[3].loadCount())
	}
}

/*************************
	Helpers
 *************************/

type testCacheParams struct {
	c         *cache
	k         *cKey
	loader    loadFunc
	validator validateFunc
	expectErr bool
	expected  entryValue
	vMatcher  types.GomegaMatcher
	msgArg    string
}

func newSuccessCacheParams(c *cache, username string, loader loadFunc, validator validateFunc,
	expected entryValue, vMatcher types.GomegaMatcher) (*testCacheParams, *cacheCounter) {

	counter := &cacheCounter{}
	return &testCacheParams{
		c:         c,
		k:         &cKey{username:   username},
		loader:    counter.countLoad(loader),
		validator: counter.countValidate(validator),
		expectErr: false,
		expected:  expected,
		vMatcher:  vMatcher,
		msgArg:    "for " + username,
	}, counter
}

func newFailedCacheParams(c *cache, username string, loader loadFunc, validator validateFunc) (*testCacheParams, *cacheCounter) {

	counter := &cacheCounter{}
	return &testCacheParams{
		c:         c,
		k:         &cKey{username:   username},
		loader:    counter.countLoad(loader),
		validator: counter.countValidate(validator),
		expectErr: true,
		msgArg:    "for " + username,
	}, counter
}

func testSuccessGetOrLoad(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, loader loadFunc, validator validateFunc,
	expected entryValue, vMatcher types.GomegaMatcher, msgArgs...interface{}) entryValue {

	v, e := c.GetOrLoad(ctx, k, loader, validator)
	g.Expect(e).To(Succeed(), fmt.Sprintf("GetOrLoad should not fail on repeated invocation %s", msgArgs...))
	g.Expect(v).To(Not(BeNil()), fmt.Sprintf("GetOrLoad should return non-nil on repeated invocation %s", msgArgs...))
	if vMatcher != nil {
		g.Expect(v).To(vMatcher, fmt.Sprintf("repeated GetOrLoad should return correct value on repeated invocation %s", msgArgs...))
	}
	if expected != nil {
		g.Expect(v).To(Equal(expected), fmt.Sprintf("GetOrLoad should return same object on repeated invocation %s", msgArgs...))
	}
	return v
}

func testFailedGetOrLoad(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, loader loadFunc, validator validateFunc,
	msgArgs...interface{}) error {

	_, e := c.GetOrLoad(ctx, k, loader, validator)
	g.Expect(e).To(Not(Succeed()), fmt.Sprintf("GetOrLoad should fail on repeated invocation %s", msgArgs...))
	return e
}

func testGetOrLoad(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, loader loadFunc, validator validateFunc,
	shouldSuccess bool, vMatcher types.GomegaMatcher, msgArgs...interface{}) (entryValue, error) {

	if shouldSuccess {
		v := testSuccessGetOrLoad(ctx, g, c, k, loader, validator, nil, vMatcher, msgArgs...)
		return v, nil
	} else {
		e := testFailedGetOrLoad(ctx, g, c, k, loader, validator, msgArgs...)
		return nil, e
	}
}

func testRepeatedGetOrLoad(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, loader loadFunc, validator validateFunc,
	n int, shouldSuccess bool, vMatcher types.GomegaMatcher, msgArgs...interface{}) (ret entryValue, err error) {

	// try to get
	for i := 0; i < n; i++ {
		newKey := *k
		if shouldSuccess {
			v := testSuccessGetOrLoad(ctx, g, c, &newKey, loader, validator, ret, vMatcher, msgArgs...)
			if ret == nil {
				ret = v
			}
		} else {
			e := testFailedGetOrLoad(ctx, g, c, &newKey, loader, validator, msgArgs...)
			if err == nil {
				err = e
			}
		}
	}
	return
}

func testRepeatedConcurrentGetOrLoad(ctx context.Context, g *gomega.WithT, params []*testCacheParams, duration time.Duration) (int, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	count := uint64(0)
	wg := &sync.WaitGroup{}

	for _, p := range params {
		go doTestRepeatedConcurrentGetOrLoad(timeoutCtx, g, p, wg, &count)
	}

	select {
	case <-timeoutCtx.Done():
		// wait for all fired goroutines to finish
		wg.Wait()
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return int(count), nil
		}
	}
	return int(atomic.LoadUint64(&count)), fmt.Errorf("concurrent test finished prematurely")
}

func doTestRepeatedConcurrentGetOrLoad(ctx context.Context, g *gomega.WithT, p *testCacheParams, wg *sync.WaitGroup, counter *uint64) {

	// try to get
	for i := 0; true; i++ {
		if ctx.Err() != nil || ctx.Err() != nil {
			break
		}

		wg.Add(1)
		newKey := *p.k
		msg := fmt.Sprintf("%s[%d]", p.msgArg, i)

		go func() {
			defer wg.Done()
			if ctx.Err() != nil || ctx.Err() != nil {
				return
			}
			atomic.AddUint64(counter, 1)
			if !p.expectErr {
				_ = testSuccessGetOrLoad(ctx, g, p.c, &newKey, p.loader, p.validator, p.expected, p.vMatcher, msg)
			} else {
				_ = testFailedGetOrLoad(ctx, g, p.c, &newKey, p.loader, p.validator, msg)
			}
		}()
	}
}

/*************************
	Loaders & Validators
 *************************/

func staticLoadFunc(loadTime time.Duration, validity time.Duration) (loadFunc, entryValue) {
	toLoad := oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Token = oauth2.NewDefaultAccessToken("Test-Token-" + utils.RandomString(10))
	})
	return func(ctx context.Context, k cKey) (v entryValue, exp time.Time, err error) {
		time.Sleep(loadTime)
		exp = time.Now().Add(validity)
		v = toLoad
		return
	}, toLoad
}

func staticErrLoadFunc(loadTime time.Duration, validity time.Duration) (loadFunc, error) {
	toErr := fmt.Errorf("synthesised error")
	return func(ctx context.Context, k cKey) (v entryValue, exp time.Time, err error) {
		time.Sleep(loadTime)
		// unhappy path valid 2 seconds
		exp = time.Now().Add(validity)
		err = toErr
		return
	}, toErr
}

func fixedValidateFunc(valid bool) validateFunc {
	return func(ctx context.Context, value entryValue) bool {
		return valid
	}
}

func notValidateFunc(not entryValue) validateFunc {
	return func(ctx context.Context, value entryValue) bool {
		//fmt.Printf("%v == %v ? %v\n", value.AccessToken().Value(), not.AccessToken().Value(), value == not)
		return value != not
	}
}

func failOccasionallyValidateFunc(n int, delay time.Duration) (validateFunc, *time.Ticker) {
	mtx := sync.Mutex{}
	var fail bool
	var count int
	fn := func() {
		mtx.Lock()
		defer mtx.Unlock()
		if count < n && !fail {
			fail = true
			count ++
		}
	}
	ticker := time.NewTicker(delay)
	go func() {
		for {
			select {
			case <-ticker.C:
				fn()
			}
		}
	}()

	return func(ctx context.Context, value entryValue) bool {
		mtx.Lock()
		defer mtx.Unlock()
		if fail {
			fail = false
			return false
		}
		return true
	}, ticker
}
