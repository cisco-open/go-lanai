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

package sd

import (
    "context"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/test"
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
		cache := NewSimpleServiceCache()
		svc := &discovery.Service{
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

func AssertCacheEntry(g *gomega.WithT, cache discovery.ServiceCache, name string, expected bool) {
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
