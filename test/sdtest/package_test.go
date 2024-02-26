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

package sdtest

import (
    "context"
    "embed"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/discovery"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "strconv"
    "testing"
)

var (
	ExpectedServices = map[string]int{
		"service1": 2,
		"service2": 1,
	}
)

//go:embed testdata/*
var testFS embed.FS

type testDI struct {
	fx.In
	Client discovery.Client
}

func TestWithMockedSDWithFileDefinition(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithMockedSD(LoadDefinition(testFS, "testdata/services.yml")),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestDiscovery(di), "TestDiscovery"),
	)
}

func TestWithMockedSDWithPropertiesDefinition(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithMockedSD(DefinitionWithPrefix("sd-mocks")),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestDiscovery(di), "TestDiscovery"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestDiscovery(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.Client).To(BeAssignableToTypeOf(&ClientMock{}), "discovery client should be *ClientMock")
		for svc, count := range ExpectedServices {
			instancer, e := di.Client.Instancer(svc)
			g.Expect(e).To(Succeed(), "getting Instancer of %s shouldn't fail", svc)
			insts, e := instancer.Instances(nil)
			g.Expect(e).To(Succeed(), "Instancer.Instances of %s shouldn't fail", svc)
			g.Expect(insts).To(HaveLen(count), "Instancer.Instances should return correct number of instances")
			for i, inst := range insts {
				expectedId := fmt.Sprintf("%s-inst-%d", svc, i)
				expectedAddr := fmt.Sprintf("192.168.0.10%d", i)
				g.Expect(inst.ID).To(Equal(expectedId), "ID should be correct of instance %d", i)
				g.Expect(inst.Address).To(Equal(expectedAddr), "Address should be correct of instance %d", i)
				g.Expect(inst.Port).To(Equal(9000 + i), "Port should be correct of instance %d", i)
				g.Expect(inst.Health).To(Equal(discovery.HealthPassing), "Health should be correct of instance %d", i)
				g.Expect(inst.Tags).To(ContainElements("mocked", "inst" + strconv.Itoa(i)), "Tags should be correct of instance %d", i)
				g.Expect(inst.Meta).To(HaveKeyWithValue("index", strconv.Itoa(i)), "Meta should be correct of instance %d", i)
			}
		}
	}
}