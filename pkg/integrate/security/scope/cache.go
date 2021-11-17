package scope

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type cKey struct {
	src        string // source username
	username   string // target username
	userId     string // target userId
	tenantExternalId string // target tenantExternalId
	tenantId   string // target tenantId
}

func (k cKey) String() string {
	user := k.username
	if user == "" {
		user = k.userId
	}
	tenant := k.tenantId
	if tenant == "" {
		tenant = k.tenantExternalId
	}
	return fmt.Sprintf("%s->%s@%s", k.src, user, tenant)
}

type entryValue oauth2.Authentication

// cEntry carries cache entry.
// after the sync.WaitGroup's Wait() func, value, expire and lastErr should be immutable
// and isLoaded() should return true
type cEntry struct {
	wg      sync.WaitGroup
	value   entryValue
	expire  time.Time
	lastErr error
	// invalid indicates whether "get" function should return it as existing entry.
	// once an entry become "invalid", it's equivalent to "not exist"
	// invalid can only be set from False to True atomically.
	// when invalid flag == 1, it's guaranteed that the entry is not valid and such status is immutable
	// when invalid flag == 0, it's NOT guaranteed the entry is "valid", goroutines should also check other fields after sync.WaitGroup's Wait()
	invalid uint64
	// loaded is used for evicting function to decide if expire is available without waiting on loader
	// because evicting func is executed periodically to act on "loaded" entries, and loaded can only be set from False to True,
	// it's not necessary to use lock to coordinate, atomic op is sufficient
	// other threads/goroutines should use sync.WaitGroup's Wait()
	loaded  uint64
}

// isExpired is NOT goroutine-safe
func (ce *cEntry) isExpired() bool {
	return !ce.expire.IsZero() && !time.Now().Before(ce.expire)
}

// isInvalidated is atomic operation
func (ce *cEntry) isInvalidated() bool {
	return atomic.LoadUint64(&ce.invalid) != 0
}

// invalidate is atomic operation
func (ce *cEntry) invalidate() {
	atomic.StoreUint64(&ce.invalid, 1)
}

// isLoaded is atomic operation
func (ce *cEntry) isLoaded() bool {
	return atomic.LoadUint64(&ce.loaded) != 0
}

// markLoaded is atomic operation
func (ce *cEntry) markLoaded() {
	atomic.StoreUint64(&ce.loaded, 1)
}

type loadFunc func(ctx context.Context, k cKey) (v entryValue, exp time.Time, err error)
type newFunc func(context.Context, *cKey) *cEntry
type validateFunc func(context.Context, entryValue) bool

type cacheOptions func(opt *cacheOption)
type cacheOption struct {
	Heartbeat time.Duration
}

type cache struct {
	mtx    sync.RWMutex
	store  map[cKey]*cEntry
	reaper *time.Ticker
}

func newCache(opts ...cacheOptions) (ret *cache) {
	opt := cacheOption{
		Heartbeat: 10 * time.Minute,
	}
	for _, fn := range opts {
		fn(&opt)
	}

	ret = &cache{
		store: map[cKey]*cEntry{},
	}
	ret.startReaper(opt.Heartbeat)
	return
}

func (c *cache) GetOrLoad(ctx context.Context, k *cKey, loader loadFunc, validator validateFunc) (entryValue, error) {
	// maxRetry should be > 0, no upper limit
	// 1. when entry exists and not expired/invalidated, no retry
	// 2. when entry is newly created, no retry
	// 3. when entry exists but expired/invalidated, mark it invalidated and retry
	const maxRetry = 2
	for i := 0; i <= maxRetry; i++ {
		// getOrNew guarantee that only one goroutine create new entry (if needed)
		// aka, getOrNew uses cache-wise RW lock to ensure such behavior
		entry, isNew := c.getOrNew(ctx, k, c.newEntryFunc(loader))
		if entry == nil {
			return nil, fmt.Errorf("[Internal Error] security Scope cache returns nil entry")
		}

		// wait for entry to load
		entry.wg.Wait()

		// from now on, entry content become immutable
		// check entry validity
		// note that we skip validation if the entry is freshly created
		if isNew || !entry.isExpired() && (entry.lastErr != nil || validator(ctx, entry.value)) {
			// valid entry
			if entry.lastErr != nil {
				return nil, entry.lastErr
			}
			return entry.value, nil
		}
		entry.invalidate()
	}

	return nil, fmt.Errorf("unable to load valid entry")
}

// newEntryFunc returns a newFunc that create an entry and kick off "loader" in separate goroutine
// this method is not goroutine safe.
func (c *cache) newEntryFunc(loader loadFunc) newFunc {
	return func(ctx context.Context, key *cKey) *cEntry {
		ret := &cEntry{}
		ret.wg.Add(1)
		// schedule load
		go c.load(ctx, key, ret, loader)
		return ret
	}
}

// load execute given loader and sent entry's sync.WaitGroup Done()
// this method is not goroutine-safe and should be invoked only once
func (c *cache) load(ctx context.Context, key *cKey, entry *cEntry, loader loadFunc) {
	v, exp, e := loader(ctx, *key)
	entry.value = v
	entry.expire = exp
	entry.lastErr = e
	entry.markLoaded()
	entry.wg.Done()
}

// getOrNew return existing entry or create and set using newIfAbsent
// this method is goroutine-safe
func (c *cache) getOrNew(ctx context.Context, pKey *cKey, newIfAbsent newFunc) (entry *cEntry, isNew bool) {
	v, ok := c.get(pKey)
	if ok {
		return v, false
	}
	return c.newIfAbsent(ctx, pKey, newIfAbsent)
}

// newIfAbsent create entry using given "creator" if the key doesn't exist. otherwise returns existing entry
// this method is goroutine-safe
func (c *cache) newIfAbsent(ctx context.Context, pKey *cKey, creator newFunc) (entry *cEntry, isNew bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if v, ok := c.getValue(pKey); ok && !v.isInvalidated() || creator == nil {
		return v, false
	}

	v := creator(ctx, pKey)
	c.setValue(pKey, v)
	return v, true
}

// set is goroutine-safe
func (c *cache) set(pKey *cKey, v *cEntry) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.setValue(pKey, v)
}

// get is goroutine-safe
func (c *cache) get(pKey *cKey) (*cEntry, bool) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	if v, ok := c.getValue(pKey); ok && !v.isInvalidated() {
		return v, ok
	}
	return nil, false
}

// getValue not goroutine-safe
func (c *cache) getValue(pKey *cKey) (*cEntry, bool) {
	if v, ok := c.store[*pKey]; ok && v != nil {
		return v, true
	}
	return nil, false
}

// setValue not goroutine-safe
func (c *cache) setValue(pKey *cKey, v *cEntry) {
	if v == nil {
		delete(c.store, *pKey)
	} else {
		c.store[*pKey] = v
		c.deleteInvalidatedValues()
	}
}

// deleteInvalidatedValues remove given keys
// this method is not goroutine-safe
func (c *cache) deleteInvalidatedValues() {
	for k, v := range c.store {
		if v.isInvalidated() {
			c.setValue(&k, nil)
		}
	}
}

func (c *cache) startReaper(interval time.Duration) {
	c.reaper = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-c.reaper.C:
				c.evict()
			}
		}
	}()
}

func (c *cache) evict() {
	// step 1, go through the store, find loaded entries (using atomic flag) and mark them invalidated if expired (with R lock)
	func() {
		c.mtx.RLock()
		defer c.mtx.RUnlock()
		for _, v := range c.store {
			if !v.isInvalidated() && v.isLoaded() && v.isExpired() {
				v.invalidate()
			}
		}
	}()

	// step 2, remove invalidated entries (with W lock)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.deleteInvalidatedValues()
}
