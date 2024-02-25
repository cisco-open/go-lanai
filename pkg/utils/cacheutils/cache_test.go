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

package cacheutils

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/cisco-open/go-lanai/test"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "github.com/onsi/gomega/types"
    "math"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

type cacheCounter struct {
	lc uint64
	vc uint64
	uc uint64
	fc uint64
}

func (c *cacheCounter) reset() {
	*c = cacheCounter{}
}

func (c *cacheCounter) countLoad(fn LoadFunc) LoadFunc {
	if fn == nil {
		return nil
	}
	return func(ctx context.Context, k Key) (interface{}, time.Time, error) {
		atomic.AddUint64(&c.lc, 1)
		return fn(ctx, k)
	}
}

func (c *cacheCounter) countValidate(fn ValidateFunc) ValidateFunc {
	if fn == nil {
		return nil
	}
	return func(ctx context.Context, v interface{}) (valid bool) {
		atomic.AddUint64(&c.vc, 1)
		defer func() {
			if !valid {
				atomic.AddUint64(&c.fc, 1)
			}
		}()
		return fn(ctx, v)
	}
}

func (c *cacheCounter) countUpdate(fn UpdateFunc) UpdateFunc {
	if fn == nil {
		return nil
	}
	return func(ctx context.Context, k Key, old interface{}) (v interface{}, exp time.Time, err error) {
		atomic.AddUint64(&c.uc, 1)
		return fn(ctx, k, old)
	}
}

func (c *cacheCounter) loadCount() int {
	return int(atomic.LoadUint64(&c.lc))
}

func (c *cacheCounter) validateCount() int {
	return int(atomic.LoadUint64(&c.vc))
}

func (c *cacheCounter) updateCount() int {
	return int(atomic.LoadUint64(&c.uc))
}

func (c *cacheCounter) invalidCount() int {
	return int(atomic.LoadUint64(&c.fc))
}

type cKey struct {
	StringKey
}


/*************************
	Tests
 *************************/
func TestCacheCorrectnessPositiveCases(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestCacheSuccessfulLoad(), "CacheSuccessfulLoad"),
		test.GomegaSubTest(SubTestCacheFailedLoad(), "CacheFailedLoad"),
		test.GomegaSubTest(SubTestCacheOnDifferentKeys(), "CacheOnDifferentKeys"),
		test.GomegaSubTest(SubTestCacheSuccessfulUpdate(), "CacheSuccessfulUpdate"),
		test.GomegaSubTest(SubTestCacheFailedUpdate(), "CacheFailedUpdate"),
		test.GomegaSubTest(SubTestCacheDelete(), "CacheDelete"),
		test.GomegaSubTest(SubTestCacheReset(), "CacheReset"),
		test.GomegaSubTest(SubTestCacheEvict(true), "CacheManualEvict"),
		test.GomegaSubTest(SubTestCacheEvict(false), "CacheAutoEvict"),
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
		c := NewMemCache()
		loader, expected := staticLoadFunc(100*time.Millisecond, 60*time.Second)

		k := cKey{"u1"}
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
		loader, expected = staticLoadFunc(100*time.Millisecond, 60*time.Second)
		validator = notValidateFunc(auth)
		_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			repeat, true, And(BeIdenticalTo(expected), Not(BeIdenticalTo(auth))), "after invalidation")

		g.Expect(counter.loadCount()).To(Equal(1), "repeated GetOrLoad should only invoke loader once")
		g.Expect(counter.validateCount()).To(Equal(repeat), "validator should be invoked once per repeated GetOrLoad invocation after invalidation")

		// try 0 valued expire time
		counter.reset()
		k = cKey{"u2"}
		loader, expected = staticNeverExpireLoadFunc(100*time.Millisecond)
		validator = fixedValidateFunc(true)
		_, _ = testGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			true, BeIdenticalTo(expected), "at first time")
		_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			repeat, true, And(BeIdenticalTo(expected), BeIdenticalTo(expected)), "after first load")

		g.Expect(counter.loadCount()).To(Equal(1), "repeated GetOrLoad should only invoke loader once")
		g.Expect(counter.validateCount()).To(Equal(repeat), "validator should be invoked once per repeated GetOrLoad invocation after invalidation")

		// try nil validator
		counter.reset()
		_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), nil,
			repeat, true, And(BeIdenticalTo(expected), BeIdenticalTo(expected)), "after first load without validator")

		g.Expect(counter.loadCount()).To(Equal(0), "repeated GetOrLoad should never invoked")
	}
}

func SubTestCacheFailedLoad() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const repeat = 5
		c := NewMemCache()
		k := cKey{"u1"}
		counter := &cacheCounter{}

		_ = testFailedGetOrLoad(ctx, g, c, &k, nil, nil, "with nil loader")

		loader, _ := staticErrLoadFunc(100*time.Millisecond, 200*time.Millisecond)
		validator := fixedValidateFunc(true)
		_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			repeat, false, nil, "after first load")

		g.Expect(counter.loadCount()).To(Equal(1), "repeated GetOrLoad should only invoke loader once")
		g.Expect(counter.validateCount()).To(Equal(0), "validator should not be invoked if loader failed")

		// try to get with previous data invalid
		counter.reset()
		time.Sleep(200 * time.Millisecond)
		loader, _ = staticErrLoadFunc(200*time.Millisecond, 200*time.Millisecond)
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
		c := NewMemCache()
		validator := fixedValidateFunc(true)

		counters := make([]*cacheCounter, keys)
		results := make([]interface{}, keys)
		for i := 0; i < keys; i++ {
			k := cKey{StringKey(fmt.Sprintf("u%d", i))}
			counters[i] = &cacheCounter{}
			loader, expected := staticLoadFunc(100*time.Millisecond, 60*time.Second)
			results[i], _ = testRepeatedGetOrLoad(ctx, g, c, &k, counters[i].countLoad(loader), counters[i].countValidate(validator),
				repeat, true, BeIdenticalTo(expected), "for "+k.String())

		}

		// assert
		for i := 0; i < keys; i++ {
			for j := i + 1; j < keys; j++ {
				g.Expect(results[i]).To(Not(BeIdenticalTo(results[j])), "different key should yield different value")
			}
			g.Expect(counters[i].loadCount()).To(Equal(1), "repeated GetOrLoad of k%d should only invoke loader once", i)
			g.Expect(counters[i].validateCount()).To(Equal(repeat-1), "validator of k%d should be invoked once per repeated GetOrLoad invocation", i)
		}
	}
}

func SubTestCacheSuccessfulUpdate() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const repeat = 5
		c := NewMemCache()
		updater, expected := staticUpdateFunc(100*time.Millisecond, 60*time.Second)
		counter := &cacheCounter{}

		k := cKey{"u1"}
		// update should do nothing before load
		testNonExistsUpdate(ctx, g, c, &k, counter.countUpdate(updater), "before loaded")
		g.Expect(counter.updateCount()).To(Equal(0), "updater should not be invoked")

		// load
		validator := fixedValidateFunc(true)
		loader, expectedInit := staticLoadFunc(0, 60*time.Second)
		_, _ = testGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			true, BeIdenticalTo(expectedInit), "at first time")

		// update should do its job
		_, _ = testUpdate(ctx, g, c, &k, counter.countUpdate(updater), expectedInit, true, BeIdenticalTo(expected), "after loaded")
		_, _ = testRepeatedUpdate(ctx, g, c, &k, counter.countUpdate(updater), expectedInit, repeat, true, BeIdenticalTo(expected))
		g.Expect(counter.updateCount()).To(Equal(repeat+1), "repeated updater should be invoked once per repeated Update invocation")
	}
}

func SubTestCacheFailedUpdate() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const repeat = 5
		c := NewMemCache()
		updater, _ := staticErrUpdateFunc(100*time.Millisecond, 60*time.Second)
		counter := &cacheCounter{}
		k := cKey{"u1"}

		// nil check
		_ = testFailedUpdate(ctx, g, c, &k, nil, nil, nil, "with nil updater")

		// update should do nothing before load
		testNonExistsUpdate(ctx, g, c, &k, counter.countUpdate(updater), "before loaded")
		g.Expect(counter.updateCount()).To(Equal(0), "updater should not be invoked")

		// load
		validator := fixedValidateFunc(true)
		loader, expectedInit := staticLoadFunc(0, 60*time.Second)
		_, _ = testGetOrLoad(ctx, g, c, &k, counter.countLoad(loader), counter.countValidate(validator),
			true, BeIdenticalTo(expectedInit), "at first time")

		// update should do its job
		_, _ = testUpdate(ctx, g, c, &k, counter.countUpdate(updater), expectedInit,
			false, BeIdenticalTo(expectedInit), "after loaded")
		_, _ = testRepeatedUpdate(ctx, g, c, &k, counter.countUpdate(updater), expectedInit,
			repeat, false, BeIdenticalTo(expectedInit))
		g.Expect(counter.updateCount()).To(Equal(repeat+1), "repeated updater should be invoked once per repeated Update invocation")
	}
}

func SubTestCacheDelete() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const longExpKeys = 3
		const repeat = 5
		c := NewMemCache(func(opt *CacheOption) {
			opt.Heartbeat = 60 * time.Second
		})
		validator := fixedValidateFunc(true)
		var toDelete cKey
		for i := 0; i < longExpKeys; i++ {
			k := cKey{StringKey(fmt.Sprintf("lu%d", i))}
			loader, expected := staticLoadFunc(0, 60*time.Second)
			_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, loader, validator,
				repeat, true, BeIdenticalTo(expected), "for "+k.String())
			toDelete = k
		}
		c.Delete(&toDelete)
		g.Expect(len(c.store)).To(Equal(2), "one entries should be removed")

		c.Delete(&cKey{"non-exists"})
		g.Expect(len(c.store)).To(Equal(2), "no entries should be removed")
	}
}

func SubTestCacheReset() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const longExpKeys = 3
		const repeat = 5
		c := NewMemCache(func(opt *CacheOption) {
			opt.Heartbeat = 60 * time.Second
		})
		validator := fixedValidateFunc(true)
		for i := 0; i < longExpKeys; i++ {
			k := cKey{StringKey(fmt.Sprintf("lu%d", i))}
			loader, expected := staticLoadFunc(0, 60*time.Second)
			_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, loader, validator,
				repeat, true, BeIdenticalTo(expected), "for "+k.String())
		}
		c.Reset()
		g.Expect(len(c.store)).To(Equal(0), "all entries should be removed")
	}
}

func SubTestCacheEvict(manual bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const longExpKeys = 3
		const shortExpKeys = 3
		const repeat = 5
		var exp = 200 * time.Millisecond
		c := NewMemCache(func(opt *CacheOption) {
			if manual {
				opt.Heartbeat = exp * 2 // make sure auto evict not in place
			} else {
				opt.Heartbeat = exp / 2 // make sure auto evict in place
			}
		})
		validator := fixedValidateFunc(true)

		for i := 0; i < shortExpKeys; i++ {
			k := cKey{StringKey(fmt.Sprintf("su%d", i))}
			loader, expected := staticLoadFunc(100*time.Millisecond, exp)
			_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, loader, validator,
				repeat, true, BeIdenticalTo(expected), "for "+k.String())
		}

		for i := 0; i < longExpKeys; i++ {
			k := cKey{StringKey(fmt.Sprintf("lu%d", i))}
			loader, expected := staticLoadFunc(100*time.Millisecond, 60*time.Second)
			_, _ = testRepeatedGetOrLoad(ctx, g, c, &k, loader, validator,
				repeat, true, BeIdenticalTo(expected), "for "+k.String())
		}

		time.Sleep(exp)
		if manual {
			c.Evict()
		}
		g.Expect(len(c.store)).To(Equal(3), "invalidated(expired) entries should be removed")
	}
}

func SubTestCacheConcurrentSoapTest() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {

		// Setup concurrent scenarios
		timeout := 1000 * time.Millisecond

		longVLoader, expectedLongV := staticLoadFunc(100*time.Millisecond, 60*time.Second)
		unstableLoader, expectedUnstable := staticLoadFunc(100*time.Millisecond, 60*time.Second)
		errLoader, _ := staticErrLoadFunc(100*time.Millisecond, 60*time.Second)
		shortVLoader, expectedShortV := staticLoadFunc(100*time.Millisecond, 200*time.Millisecond)

		stableValidator := fixedValidateFunc(true)
		unstableValidator, ticker := failOccasionallyValidateFunc(3, 250*time.Millisecond)
		defer ticker.Stop()

		longVUpdater := copyUpdateFunc(0, 60*time.Second)
		unstableUpdater := copyUpdateFunc(0, 60*time.Second)
		errUpdater, _ := staticErrUpdateFunc(0, 60*time.Second)

		c := NewMemCache(func(opt *CacheOption) {
			opt.Heartbeat = 100 * time.Millisecond
		})
		params := make([]*testCacheParams, 4)
		counters := make([]*cacheCounter, 4)
		params[0], counters[0] = newSuccessCacheParams(c, "long-validity-user", longVLoader, stableValidator, longVUpdater, expectedLongV, nil)
		params[1], counters[1] = newSuccessCacheParams(c, "unstable-user", unstableLoader, unstableValidator, unstableUpdater, expectedUnstable, nil)
		params[2], counters[2] = newFailedCacheParams(c, "non-exist-user", errLoader, stableValidator, errUpdater)
		params[3], counters[3] = newSuccessCacheParams(c, "short-validity-user", shortVLoader, stableValidator, nil, expectedShortV, nil)

		// first, trigger GetOrLoad with special key and short exp period (for evict)
		_, _ = testGetOrLoad(ctx, g, c, &cKey{"to-evicted-user"}, shortVLoader, stableValidator,
			true, BeIdenticalTo(expectedShortV), " for to-evicted-user")

		// Run concurrent test
		count, e := testRepeatedConcurrentOperations(ctx, g, params, timeout)
		if e != nil {
			t.Errorf("%v", e)
		}

		// Assert invocation count
		fmt.Printf("load counts: %d %d %d %d\n", counters[0].loadCount(), counters[1].loadCount(), counters[2].loadCount(), counters[3].loadCount())
		fmt.Printf("invalid counts: %d %d %d %d\n", counters[0].invalidCount(), counters[1].invalidCount(), counters[2].invalidCount(), counters[3].invalidCount())
		g.Expect(count).To(BeNumerically(">", 500), "GetOrLoad should be invoked many times")

		// long stable
		g.Expect(counters[0].loadCount()).To(Equal(1), "repeated GetOrLoad of long validity should only invoke loader once")
		// long unstable
		maxLoad := counters[1].invalidCount() + 1 // actual invalid count + initial load
		minLoad := int(math.Min(float64(maxLoad), 2))
		g.Expect(counters[1].loadCount()).To(BeNumerically(">=", minLoad), "repeated GetOrLoad of unstable result should invoke loader more than once if validator returns false at least once")
		g.Expect(counters[1].loadCount()).To(BeNumerically("<=", maxLoad), "repeated GetOrLoad of unstable result should invoke loader %d times (invalid count + initial load)", maxLoad)
		// error
		g.Expect(counters[2].loadCount()).To(Equal(1), "repeated GetOrLoad of error result should invoke loader once")
		g.Expect(counters[2].validateCount()).To(Equal(0), "repeated GetOrLoad of error result should never invoke validator ")
		// short stable
		g.Expect(counters[3].loadCount()).To(BeNumerically(">=", 2), "repeated GetOrLoad of short validity should invoke loader more than once")
		g.Expect(counters[3].loadCount()).To(BeNumerically("<=", 10), "repeated GetOrLoad of short validity should not invoke loader too many times ( <= 10)")

		g.Expect(len(c.store)).To(BeNumerically("<=", 4), "invalidated(expired) entries should be removed")
	}
}

/*************************
	Helpers
 *************************/

func testSuccessGetOrLoad(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, loader LoadFunc, validator ValidateFunc,
	expected interface{}, vMatcher types.GomegaMatcher, msgArgs ...interface{}) interface{} {

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
	c *cache, k *cKey, loader LoadFunc, validator ValidateFunc,
	msgArgs ...interface{}) error {

	_, e := c.GetOrLoad(ctx, k, loader, validator)
	g.Expect(e).To(Not(Succeed()), fmt.Sprintf("GetOrLoad should fail on repeated invocation %s", msgArgs...))
	return e
}

func testGetOrLoad(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, loader LoadFunc, validator ValidateFunc,
	shouldSuccess bool, vMatcher types.GomegaMatcher, msgArgs ...interface{}) (interface{}, error) {

	if shouldSuccess {
		v := testSuccessGetOrLoad(ctx, g, c, k, loader, validator, nil, vMatcher, msgArgs...)
		return v, nil
	} else {
		e := testFailedGetOrLoad(ctx, g, c, k, loader, validator, msgArgs...)
		return nil, e
	}
}

func testRepeatedGetOrLoad(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, loader LoadFunc, validator ValidateFunc,
	n int, shouldSuccess bool, vMatcher types.GomegaMatcher, msgArgs ...interface{}) (ret interface{}, err error) {

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

func testSuccessUpdate(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, updater UpdateFunc,
	original interface{}, vMatcher types.GomegaMatcher, msgArgs ...interface{}) interface{} {

	ok, e := c.Update(ctx, k, updater)
	g.Expect(e).To(Succeed(), fmt.Sprintf("Update should not fail on repeated invocation %s", msgArgs...))

	var v interface{}
	if entry, exists := c.get(k); exists && entry != nil {
		v = entry.value
	}
	if ok && vMatcher != nil {
		g.Expect(v).To(vMatcher, fmt.Sprintf("current value should be correct after Update %s", msgArgs...))
	}
	if ok && original != nil {
		g.Expect(v).To(Not(Equal(original)), fmt.Sprintf("current value should not be same object on repeated Update invocation %s", msgArgs...))
	}
	return v
}

func testFailedUpdate(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, updater UpdateFunc,
	original interface{}, vMatcher types.GomegaMatcher, msgArgs ...interface{}) error {

	ok, e := c.Update(ctx, k, updater)
	if ok {
		g.Expect(e).To(Not(Succeed()), fmt.Sprintf("Update should fail on repeated invocation %s", msgArgs...))
	}

	var v interface{}
	if entry, exists := c.get(k); exists && entry != nil {
		v = entry.value
	}
	if ok && vMatcher != nil {
		g.Expect(v).To(vMatcher, fmt.Sprintf("current value should not be changed after failed Update %s", msgArgs...))
	}
	if ok && original != nil {
		g.Expect(v).To(Equal(original), fmt.Sprintf("current value should not be changed after failed on repeated Update invocation %s", msgArgs...))
	}
	return e
}

func testNonExistsUpdate(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, updater UpdateFunc, msgArgs ...interface{}) {

	ok, e := c.Update(ctx, k, updater)
	g.Expect(ok).To(BeFalse(), fmt.Sprintf("Update should not do anything on repeated invocation %s", msgArgs...))
	g.Expect(e).To(Succeed(), fmt.Sprintf("Update should not fail on repeated invocation %s", msgArgs...))
}

func testUpdate(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, updater UpdateFunc, original interface{},
	shouldSuccess bool, vMatcher types.GomegaMatcher, msgArgs ...interface{}) (interface{}, error) {

	if shouldSuccess {
		v := testSuccessUpdate(ctx, g, c, k, updater, original, vMatcher, msgArgs...)
		return v, nil
	} else {
		e := testFailedUpdate(ctx, g, c, k, updater, original, vMatcher, msgArgs...)
		return nil, e
	}
}

func testRepeatedUpdate(ctx context.Context, g *gomega.WithT,
	c *cache, k *cKey, updater UpdateFunc, original interface{},
	n int, shouldSuccess bool, vMatcher types.GomegaMatcher, msgArgs ...interface{}) (ret interface{}, err error) {

	// try to get
	for i := 0; i < n; i++ {
		newKey := *k
		if shouldSuccess {
			v := testSuccessUpdate(ctx, g, c, &newKey, updater, original, vMatcher, msgArgs...)
			if ret == nil {
				ret = v
			}
		} else {
			e := testFailedUpdate(ctx, g, c, &newKey, updater, original, vMatcher, msgArgs...)
			if err == nil {
				err = e
			}
		}
	}
	return
}

/*************************
	Concurrent Helpers
 *************************/

type testCacheParams struct {
	c         *cache
	k         *cKey
	loader    LoadFunc
	validator ValidateFunc
	updater   UpdateFunc
	expectErr bool
	expected  interface{}
	vMatcher  types.GomegaMatcher
	msgArg    string
}

func newSuccessCacheParams(c *cache, key string, loader LoadFunc, validator ValidateFunc, updater UpdateFunc,
	expected interface{}, vMatcher types.GomegaMatcher) (*testCacheParams, *cacheCounter) {

	counter := &cacheCounter{}
	return &testCacheParams{
		c:         c,
		k:         &cKey{StringKey(key)},
		loader:    counter.countLoad(loader),
		validator: counter.countValidate(validator),
		updater:   counter.countUpdate(updater),
		expectErr: false,
		expected:  expected,
		vMatcher:  vMatcher,
		msgArg:    "for " + key,
	}, counter
}

func newFailedCacheParams(c *cache, key string, loader LoadFunc, validator ValidateFunc, updater UpdateFunc) (*testCacheParams, *cacheCounter) {

	counter := &cacheCounter{}
	return &testCacheParams{
		c:         c,
		k:         &cKey{StringKey(key)},
		loader:    counter.countLoad(loader),
		validator: counter.countValidate(validator),
		updater:   counter.countUpdate(updater),
		expectErr: true,
		msgArg:    "for " + key,
	}, counter
}

func testRepeatedConcurrentOperations(ctx context.Context, g *gomega.WithT, params []*testCacheParams, duration time.Duration) (int, error) {

	timeoutCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	count := uint64(0)
	wg := &sync.WaitGroup{}

	for _, p := range params {
		if p == nil {
			continue
		}
		go doTestRepeatedConcurrentOperations(timeoutCtx, g, p, wg, &count)
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

func doTestRepeatedConcurrentOperations(ctx context.Context, g *gomega.WithT, p *testCacheParams, wg *sync.WaitGroup, counter *uint64) {
	// try to get
	for i := 0; true; i++ {
		if ctx.Err() != nil || ctx.Err() != nil {
			break
		}

		wg.Add(1)
		newKey := *p.k
		msg := fmt.Sprintf("%s[%d]", p.msgArg, i)

		// load
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
		// update
		if p.updater != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if ctx.Err() != nil || ctx.Err() != nil {
					return
				}
				atomic.AddUint64(counter, 1)
				// Note: relaxed condition, we don't check against original object in concurrent test, coz we are using copy updater
				if !p.expectErr {
					_ = testSuccessUpdate(ctx, g, p.c, &newKey, p.updater, nil, p.vMatcher, msg)
				} else {
					_ = testFailedUpdate(ctx, g, p.c, &newKey, p.updater, nil, p.vMatcher, msg)
				}
			}()
		}
	}
}


/***********************************
	Loaders, Validators & Updaters
 ***********************************/

type CachedTestValue struct {
	Value   string
}

func staticLoadFunc(loadTime time.Duration, validity time.Duration) (LoadFunc, interface{}) {
	toLoad := CachedTestValue{
		Value: "Test-Value-" + utils.RandomString(10),
	}
	return func(ctx context.Context, k Key) (v interface{}, exp time.Time, err error) {
		time.Sleep(loadTime)
		exp =  time.Now().Add(validity)
		v = toLoad
		return
	}, toLoad
}

func staticNeverExpireLoadFunc(loadTime time.Duration) (LoadFunc, interface{}) {
	toLoad := CachedTestValue{
		Value: "Test-Value-" + utils.RandomString(10),
	}
	return func(ctx context.Context, k Key) (v interface{}, exp time.Time, err error) {
		time.Sleep(loadTime)
		exp =  time.Time{}
		v = toLoad
		return
	}, toLoad
}

func staticErrLoadFunc(loadTime time.Duration, validity time.Duration) (LoadFunc, error) {
	toErr := fmt.Errorf("synthesised error")
	return func(ctx context.Context, k Key) (v interface{}, exp time.Time, err error) {
		time.Sleep(loadTime)
		exp = time.Now().Add(validity)
		err = toErr
		return
	}, toErr
}

func fixedValidateFunc(valid bool) ValidateFunc {
	return func(ctx context.Context, value interface{}) bool {
		return valid
	}
}

func notValidateFunc(not interface{}) ValidateFunc {
	return func(ctx context.Context, value interface{}) bool {
		//fmt.Printf("%v == %v ? %v\n", value.AccessToken().Value(), not.AccessToken().Value(), value == not)
		return value != not
	}
}

func failOccasionallyValidateFunc(maxN int, delay time.Duration) (ValidateFunc, *time.Ticker) {
	mtx := sync.Mutex{}
	var fail bool
	var count int
	fn := func() {
		mtx.Lock()
		defer mtx.Unlock()
		if !fail {
			fail = true

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

	return func(ctx context.Context, value interface{}) bool {
		mtx.Lock()
		defer mtx.Unlock()
		if count < maxN && fail {
			fail = false
			count++
			return false
		}
		return true
	}, ticker
}

func staticUpdateFunc(updateTime time.Duration, validity time.Duration) (UpdateFunc, interface{}) {
	toLoad := CachedTestValue{
		Value: "Test-Value-" + utils.RandomString(10),
	}
	return func(ctx context.Context, k Key, old interface{}) (v interface{}, exp time.Time, err error) {
		time.Sleep(updateTime)
		exp = time.Now().Add(validity)
		v = toLoad
		return
	}, toLoad
}

func staticErrUpdateFunc(updateTime time.Duration, validity time.Duration) (UpdateFunc, error) {
	toErr := fmt.Errorf("synthesised error")
	return func(ctx context.Context, k Key, old interface{}) (v interface{}, exp time.Time, err error) {
		time.Sleep(updateTime)
		exp = time.Now().Add(validity)
		err = toErr
		v = old
		return
	}, toErr
}

func copyUpdateFunc(updateTime time.Duration, validity time.Duration) UpdateFunc {
	return func(ctx context.Context, k Key, old interface{}) (v interface{}, exp time.Time, err error) {
		time.Sleep(updateTime)
		exp = time.Now().Add(validity + updateTime)
		v = old
		return
	}
}
