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

package consul_test

import (
    "context"
    "github.com/cisco-open/go-lanai/pkg/consul"
    consulinit "github.com/cisco-open/go-lanai/pkg/consul/init"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/ittest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "gopkg.in/dnaeon/go-vcr.v3/recorder"
    "testing"
)

/*************************
	Tests Setup Helpers
 *************************/

func RecordedConsulProvider() fx.Annotated {
    return fx.Annotated{
        Group:  "consul",
        Target: ConsulWithRecorder,
    }
}

func ConsulWithRecorder(recorder *recorder.Recorder) consul.Options {
    return func(cfg *consul.ClientConfig) error {
        switch {
        case cfg.Transport != nil:
            cfg.HttpClient = recorder.GetDefaultClient()
        case cfg.HttpClient != nil:
            if cfg.HttpClient.Transport != nil {
                recorder.SetRealTransport(cfg.HttpClient.Transport)
            }
            cfg.HttpClient.Transport = recorder
        default:
            cfg.HttpClient = recorder.GetDefaultClient()
        }

        return nil
    }
}

//func NewTestClient(g *gomega.WithT, props consul.ConnectionProperties, recorder *recorder.Recorder) (*consul.Connection,) {
//    client, e := consul.New(consul.WithProperties(props), ConsulWithRecorder(recorder))
//    g.Expect(e).To(Succeed(), "create consul client should not fail")
//    return client
//}


/*************************
	Tests
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		ittest.PackageHttpRecordingMode(),
//	)
//}

type TestConnDI struct {
    fx.In
    Client *consul.Connection
}

func TestConsulConnection(t *testing.T) {
    di := TestConnDI{}
    test.RunTest(context.Background(), t,
        apptest.Bootstrap(),
        ittest.WithHttpPlayback(t),
        apptest.WithModules(consulinit.Module),
        apptest.WithFxOptions(
            fx.Provide(RecordedConsulProvider()),
        ),
        apptest.WithDI(&di),
        test.GomegaSubTest(SubTestKVOperations(&di), "TestKVOperations"),
    )
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestKVOperations(di *TestConnDI) test.GomegaSubTestFunc {
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