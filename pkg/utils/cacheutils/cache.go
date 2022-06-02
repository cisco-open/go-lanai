package cacheutils

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type Key interface{}

type MemCache interface {
	// GetOrLoad try to get cached entry, using provided validator to check the entry, if not valid, try to load it.
	// If there is any error during load, it's cached and returned from this method
	// Note: this method is the main method for this in-memory cache
	GetOrLoad(ctx context.Context, k Key, loader LoadFunc, validator ValidateFunc) (interface{}, error)
	// Update Utility method, force change the loaded entry's value.
	// If given key doesn't exist or is invalidated, this function does nothing and return false, otherwise returns true.
	// If there is any error during update, it's cached and returned from this method
	// If this Update is used while the entry is loading, it will wait until loading finishes and then perform update
	Update(ctx context.Context, k Key, updater UpdateFunc) (bool, error)
	// Delete Utility method, remove the cached entry of given key, regardless if it's valid
	Delete(k Key)
	// Reset Utility method, reset the cache and remove all entries, regardless if they are valid
	Reset()
	// Evict Utility method, cleanup the cache, removing any invalid entries, free up memory
	// Note: this process is also performed periodically, normally there is no need to call this function manually
	Evict()
}

type LoadFunc func(ctx context.Context, k Key) (v interface{}, exp time.Time, err error)
type UpdateFunc func(ctx context.Context, k Key, old interface{}) (v interface{}, exp time.Time, err error)
type ValidateFunc func(ctx context.Context, v interface{}) bool

type CacheOptions func(opt *CacheOption)
type CacheOption struct {
	Heartbeat time.Duration
	LoadRetry int
}

// cEntry carries cache entry.
// after the sync.WaitGroup's Wait() func, value, expire and lastErr should be immutable
// and isLoaded() should return true
type cEntry struct {
	wg      sync.WaitGroup
	value   interface{}
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
	loaded uint64
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

type newEntryFunc func(ctx context.Context, k Key) *cEntry

type replaceEntryFunc func(ctx context.Context, k Key, old *cEntry) *cEntry

type cache struct {
	CacheOption
	mtx    sync.RWMutex
	store  map[Key]*cEntry
	reaper *time.Ticker
}

func NewMemCache(opts ...CacheOptions) *cache {
	opt := CacheOption{
		Heartbeat: 10 * time.Minute,
		LoadRetry: 2,
	}
	for _, fn := range opts {
		fn(&opt)
	}

	c := &cache{
		CacheOption: opt,
		store:       map[Key]*cEntry{},
	}
	c.startReaper()
	return c
}

func (c *cache) GetOrLoad(ctx context.Context, k Key, loader LoadFunc, validator ValidateFunc) (interface{}, error) {
	if loader == nil {
		return nil, fmt.Errorf("unable to load valid entry: LoadFunc is nil")
	}
	// maxRetry should be > 0, no upper limit
	// 1. when entry exists and not expired/invalidated, no retry
	// 2. when entry is newly created, no retry
	// 3. when entry exists but expired/invalidated, mark it invalidated and retry
	for i := 0; i <= c.LoadRetry; i++ {
		// getOrNew guarantee that only one goroutine create new entry (if needed)
		// aka, getOrNew uses cache-wise RW lock to ensure such behavior
		entry, isNew := c.getOrNew(ctx, k, c.newEntryFunc(loader))
		if entry == nil {
			return nil, fmt.Errorf("[Internal Error] cache returns nil entry")
		}

		// wait for entry to load
		entry.wg.Wait()

		// from now on, entry content become immutable
		// check entry validity
		// note that we skip validation if the entry is freshly created
		if isNew || !entry.isExpired() && (entry.lastErr != nil || validator != nil && validator(ctx, entry.value)) {
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

func (c *cache) Update(ctx context.Context, k Key, updater UpdateFunc) (bool, error) {
	if updater == nil {
		return false, fmt.Errorf("unable to update: UpdateFunc is nil")
	}

	existing, ok := c.get(k)
	if !ok {
		return false, nil
	}
	existing.wg.Wait()

	newEntry := c.replaceIfPresent(ctx, k, c.updateEntryFunc(updater))
	if newEntry == nil {
		return false, nil
	}
	newEntry.wg.Wait()

	if !newEntry.isExpired() {
		return true, newEntry.lastErr
	}
	return true, nil
}

func (c *cache) Delete(k Key) {
	c.set(k, nil)
}

func (c *cache) Reset() {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	c.store = map[Key]*cEntry{}
}

func (c *cache) Evict() {
	c.evict()
}

// newEntryFunc returns a newEntryFunc that create an entry and kick off "loader" in separate goroutine
// this method is not goroutine safe.
func (c *cache) newEntryFunc(loader LoadFunc) newEntryFunc {
	return func(ctx context.Context, key Key) *cEntry {
		ret := &cEntry{}
		ret.wg.Add(1)
		// schedule load
		go c.load(ctx, key, ret, loader)
		return ret
	}
}

// updateEntryFunc returns a replaceEntryFunc that create an entry using given UpdateFunc and old entry
// this method is not goroutine safe.
// this method assume the old entry is not nil and already loaded
func (c *cache) updateEntryFunc(updater UpdateFunc) replaceEntryFunc {
	return func(ctx context.Context, key Key, old *cEntry) *cEntry {
		ret := &cEntry{
			value:   old.value,
			expire:  old.expire,
			lastErr: old.lastErr,
		}
		ret.wg.Add(1)
		// wrap updater as loader
		go c.load(ctx, key, ret, func(ctx context.Context, k Key) (interface{}, time.Time, error) {
			return updater(ctx, k, old.value)
		})
		return ret
	}
}

// load execute given loader and sent entry's sync.WaitGroup Done()
// this method is not goroutine-safe and should be invoked only once
func (c *cache) load(ctx context.Context, key Key, entry *cEntry, loader LoadFunc) {
	v, exp, e := loader(ctx, key)
	entry.value = v
	entry.expire = exp
	entry.lastErr = e
	entry.markLoaded()
	entry.wg.Done()
}

// getOrNew return existing entry or create and set using newIfAbsent
// this method is goroutine-safe
func (c *cache) getOrNew(ctx context.Context, pKey Key, newIfAbsent newEntryFunc) (entry *cEntry, isNew bool) {
	v, ok := c.get(pKey)
	if ok {
		return v, false
	}
	return c.newIfAbsent(ctx, pKey, newIfAbsent)
}

// newIfAbsent create entry using given "creator" if the key doesn't exist. otherwise returns existing entry
// this method is goroutine-safe
func (c *cache) newIfAbsent(ctx context.Context, pKey Key, creator newEntryFunc) (entry *cEntry, isNew bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if v, ok := c.getValue(pKey); ok && !v.isInvalidated() {
		return v, false
	}

	v := creator(ctx, pKey)
	c.setValue(pKey, v)
	return v, true
}

// replaceIfPresent create entry using given "replacer" and replace the current entry if the key exists. otherwise returns nil
// this method is goroutine-safe
func (c *cache) replaceIfPresent(ctx context.Context, pKey Key, replacer replaceEntryFunc) (entry *cEntry) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	existing, ok := c.getValue(pKey)
	if !ok || existing.isInvalidated() {
		return nil
	}

	existing.wg.Wait()
	entry = replacer(ctx, pKey, existing)
	c.setValue(pKey, entry)
	return entry
}

// set is goroutine-safe
func (c *cache) set(pKey Key, v *cEntry) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.setValue(pKey, v)
}

// get is goroutine-safe
func (c *cache) get(pKey Key) (*cEntry, bool) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	if v, ok := c.getValue(pKey); ok && !v.isInvalidated() {
		return v, ok
	}
	return nil, false
}

// getValue not goroutine-safe
func (c *cache) getValue(pKey Key) (*cEntry, bool) {
	k := reflect.Indirect(reflect.ValueOf(pKey)).Interface()
	if v, ok := c.store[k]; ok && v != nil {
		return v, true
	}
	return nil, false
}

// setValue not goroutine-safe
func (c *cache) setValue(pKey Key, v *cEntry) {
	k := reflect.Indirect(reflect.ValueOf(pKey)).Interface()
	if v == nil {
		delete(c.store, k)
	} else {
		c.store[k] = v
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

func (c *cache) startReaper() {
	c.reaper = time.NewTicker(c.Heartbeat)
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
