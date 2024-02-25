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

package dsyncmock

import (
	"context"
	"github.com/cisco-open/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSyncManager(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestNoopSyncManager(), "TestNoopSyncManager"),
	)
}

func SubTestNoopSyncManager() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		out := ProvideNoopSyncManager()
		manager := out.TestSyncManager
		l, e := manager.Lock("test-key")
		g.Expect(e).To(Succeed())
		g.Expect(l.Key()).To(BeEquivalentTo("test-key"))
		g.Expect(l.Lock(ctx)).To(Succeed())
		g.Expect(l.TryLock(ctx)).To(Succeed())
		g.Expect(l.Release()).To(Succeed())
		g.Expect(l.Lost()).To(HaveLen(0))
	}
}
