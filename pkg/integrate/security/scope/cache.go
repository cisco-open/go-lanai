package scope

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
	"sync"
	"time"
)

type cKey struct {
	src        string // source username
	username   string // target username
	userId     string // target userId
	tenantName string // target tenantName
	tenantId   string // target tenantId
}

func (k cKey) String() string {
	user := k.username
	if user == "" {
		user = k.userId
	}
	tenant := k.tenantId
	if tenant == "" {
		tenant = k.tenantName
	}
	return fmt.Sprintf("%s->%s@%s", k.src, user, tenant)
}

// cEntry carries cache entry.
// All methods are NOT goroutine-safe
// All fields should be immutable after the sync.WaitGroup's Wait() func
type cEntry struct {
	wg      sync.WaitGroup
	value   entryValue
	expire  time.Time
	lastErr error
}

type entryValue oauth2.Authentication

func (ce *cEntry) isValid() bool {
	return ce.expire.IsZero() || time.Now().Before(ce.expire)
}

type loadFunc func(ctx context.Context, k cKey) (v entryValue, exp time.Time, err error)
type newFunc func(context.Context, *cKey) *cEntry

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


func (c *cache) GetOrLoad(ctx context.Context, k *cKey, loader loadFunc) (entryValue, error) {
	fmt.Printf("getting key-%v...\n", k)
	entry := c.getOrNew(ctx, k, c.newEntryFunc(loader))
	if entry == nil {
		return nil, fmt.Errorf("[Internal Error] security Scope cache returns nil entry")
	}

	// wait for entry to load
	entry.wg.Wait()
	fmt.Printf("got key-%v=%v, err=%v\n", k, entry.value, entry.lastErr)

	// from now on, entry become immutable
	if !entry.isValid() {
		return nil, fmt.Errorf("[Internal Error] security Scope cache returns invalid entry")
	}

	if entry.lastErr != nil {
		return nil, entry.lastErr
	}
	return entry.value, nil
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
	entry.wg.Done()
}

// getOrNew return existing entry or create and set using newIfAbsent
// this method is goroutine-safe
func (c *cache) getOrNew(ctx context.Context, pKey *cKey, newIfAbsent newFunc) *cEntry {
	c.mtx.RLock()
	v, ok := c.getValue(pKey)
	c.mtx.RUnlock()
	if ok {
		return v
	}

	return c.setIfAbsent(ctx, pKey, newIfAbsent)
}

// setIfAbsent set entry using given "creator" if the key doesn't exist. otherwise returns existing entry
// this method is goroutine-safe
func (c *cache) setIfAbsent(ctx context.Context, pKey *cKey, creator func(context.Context, *cKey) *cEntry) *cEntry {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if v, ok := c.getValue(pKey); ok || creator == nil {
		return v
	}

	v := creator(ctx, pKey)
	c.setValue(pKey, v)
	return v
}

// getValue not goroutine-safe
func (c *cache) getValue(pKey *cKey) (*cEntry, bool) {
	if v, ok := c.store[*pKey]; ok && v != nil && v.isValid() {
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
		c.tryEvict()
	}
}

func (c *cache) startReaper(interval time.Duration) {
	c.reaper = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-c.reaper.C:
				c.tryEvict()
			}
		}
	}()
}

// tryEvict remove invalid and expired entries
// this method is not goroutine-safe
func (c *cache) tryEvict() {
	now := time.Now()
	for k, v := range c.store {
		if !v.isValid() || !v.expire.IsZero() && !now.Before(v.expire) {
			fmt.Printf("evicting key-%v...\n", k)
			delete(c.store, k)
		}
	}
}
