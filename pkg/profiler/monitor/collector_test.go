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

package monitor

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/redis"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/embedded"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

/*************************
	Setup
 *************************/

/*************************
	Test
 *************************/

type testDI struct {
	fx.In
	DataCollector *dataCollector
}

func TestDataCollector(t *testing.T) {
	SamplingRate = 10 * time.Millisecond
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		embedded.WithRedis(),
		apptest.WithModules(Module, redis.Module),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestSubscribe(di), "TestSubscribe"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestSubscribe(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ch, id, e := di.DataCollector.Subscribe()
		g.Expect(e).To(Succeed())
		g.Expect(id).To(Not(BeEmpty()))
		g.Expect(ch).To(Not(BeNil()))

		for i := 0; i < 10; i++ {
			select {
			case f := <-ch:
				g.Expect(f).To(Not(BeZero()))
			case <-ctx.Done():
			}
		}
		di.DataCollector.Unsubscribe(id)
	}
}
