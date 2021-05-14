package discovery

import (
	"time"
)

// simpleServiceCache implements ServiceCache with map[string]*Service as a back storage
// simpleServiceCache is not goroutine-safe
type simpleServiceCache struct {
	cache map[string]*Service
	exp   map[string]time.Time
}

func newSimpleServiceCache() ServiceCache {
	// prepare cache
	return &simpleServiceCache{
		cache: map[string]*Service{},
		exp:   map[string]time.Time{},
	}
}

func (c *simpleServiceCache) Get(key string) *Service {
	if c.isExpired(key, time.Now()) {
		return nil
	}
	return c.cache[key]
}

func (c *simpleServiceCache) Set(key string, svc *Service) *Service {
	existing := c.Get(key)
	c.cache[key] = svc
	return existing
}


func (c *simpleServiceCache) SetWithTTL(key string, svc *Service, ttl time.Duration) *Service {
	if ttl <= 0 {
		return c.Set(key, svc)
	}
	existing := c.Get(key)
	c.cache[key] = svc
	c.exp[key] = time.Now().Add(ttl)
	return existing
}

func (c *simpleServiceCache) Has(key string) bool {
	v := c.Get(key)
	return v != nil
}

func (c *simpleServiceCache) Entries() map[string]*Service {
	ret := make(map[string]*Service)
	now := time.Now()
	for k, v := range c.cache {
		if c.isExpired(k, now) {
			continue
		}
		ret[k] = v
	}
	return ret
}

func (c *simpleServiceCache) isExpired(key string, now time.Time) bool {
	exp, ok := c.exp[key]
	if ok && exp.Before(now) {
		delete(c.cache, key)
		delete(c.exp, key)
		return true
	}
	return false
}

/***********************
	Helpers
 ***********************/


