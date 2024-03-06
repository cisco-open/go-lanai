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

package dnssd_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/discovery/dnssd"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Tests
 *************************/

type TestModuleDI struct {
	fx.In
	AppContext      *bootstrap.ApplicationContext
	DiscoveryClient discovery.Client
}

func TestModuleInit(t *testing.T) {
	di := TestModuleDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithDNSServer(),
		apptest.WithModules(dnssd.Module),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestServices()),
		test.SubTestTeardown(TeardownTestServices()),
		test.GomegaSubTest(SubTestDiscoveryClient(&di), "TestDiscoveryClient"),
		//test.GomegaSubTest(SubTestWithoutProto(&di), "TestWithDefaultSelector"),
		//test.GomegaSubTest(SubTestWithServiceUpdates(&di), "TestWithServiceUpdates"),
		//test.GomegaSubTest(SubTestWithGoKitCompatibility(&di), "TestWithGoKitCompatibility"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestDiscoveryClient(di *TestModuleDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.DiscoveryClient).To(BeAssignableToTypeOf(dnssd.NewDiscoveryClient(ctx)))

		// get instancer
		instancer, e := di.DiscoveryClient.Instancer(ServiceName1)
		g.Expect(e).To(Succeed(), "getting instancer should not fail")
		g.Expect(instancer).ToNot(BeNil(), "instancer should not be nil")
		g.Expect(instancer.ServiceName()).To(Equal(ServiceName1), "instancer's service name should be correct")

		// via service
		svc := instancer.Service()
		g.Expect(svc).ToNot(BeNil(), "instancer should return non-nil service")
		g.Expect(svc.Insts).To(HaveLen(2), "instancer should return services with all matching instances")

		// without additional selector
		TryInstancerWithMatcher(g, instancer, nil, []*MockedService{
			&MockedServices[ServiceName1][0], &MockedServices[ServiceName1][1],
		})
	}
}

/*************************
	Helpers
 *************************/
