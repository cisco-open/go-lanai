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

package consultest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Tests
 *************************/

type TestConsulDI struct {
	fx.In
	Client *consul.Connection
}

func TestConsulConnection(t *testing.T) {
	di := TestConsulDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t,
			//HttpRecordingMode(),
			MoreHTTPVCROptions(ittest.HttpRecordName("TestBasicOperations")),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestKVOperations(&di), "TestKVOperations"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestKVOperations(di *TestConsulDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const testKVPath = "test/new-value"
		e := di.Client.SetKeyValue(ctx, testKVPath, []byte("good"))
		g.Expect(e).To(Succeed(), "set KV should not fail")

		v, e := di.Client.GetKeyValue(ctx, testKVPath)
		g.Expect(e).To(Succeed(), "get KV should not fail")
		g.Expect(v).To(BeEquivalentTo("good"), "get KV should have correct result")

		rs, e := di.Client.ListKeyValuePairs(ctx, "test")
		g.Expect(e).To(Succeed(), "list KVs should not fail")
		g.Expect(rs).To(HaveKeyWithValue("new-value", "good"), "list KVs should have correct result")
	}
}
