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

// NewSimpleServiceCache returns a ServiceCache with map[string]*Service as a back storage
// This ServiceCache is not goroutine-safe
func NewSimpleServiceCache() ServiceCache {
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


