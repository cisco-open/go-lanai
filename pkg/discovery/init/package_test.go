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

package discovery_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	discoveryinit "github.com/cisco-open/go-lanai/pkg/discovery/init"
	"github.com/cisco-open/go-lanai/pkg/discovery/init/testdata"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/consultest"
	"github.com/cisco-open/go-lanai/test/ittest"
	"github.com/hashicorp/consul/api"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

const TestRegisterFuzzyJsonPathTags = `$.Tags`
const TestServiceID = `testservice-8080-d8755f792d`

/*************************
	Test Setup
 *************************/

func OverrideTestServiceID(customizers *discovery.Customizers) {
	customizers.Add(discovery.CustomizerFunc(func(ctx context.Context, reg *api.AgentServiceRegistration) {
		reg.ID = TestServiceID
		reg.Tags = append(reg.Tags, TestServiceID)
	}))
}

/*************************
	Tests
 *************************/

type TestModuleDI struct {
	fx.In
	Consul              *consul.Connection
	AppContext          *bootstrap.ApplicationContext
	DiscoveryProperties discovery.DiscoveryProperties
	DiscoveryClient     discovery.Client
	Registration        *api.AgentServiceRegistration
	Customizers         *discovery.Customizers
}

func TestModuleInit(t *testing.T) {
	di := TestModuleDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		consultest.WithHttpPlayback(t,
			//consultest.HttpRecordingMode(),
			// Note: tags may contains build time, should be ignored
			consultest.MoreHTTPVCROptions(ittest.HttpRecordMatching(ittest.FuzzyJsonPaths(
				TestRegisterFuzzyJsonPathTags,
			))),
		),
		apptest.WithBootstrapConfigFS(testdata.TestBootstrapFS),
		apptest.WithConfigFS(testdata.TestApplicationFS),
		apptest.WithModules(discoveryinit.Module),
		apptest.WithFxOptions(
			fx.Invoke(OverrideTestServiceID),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestVerifyRegistration(&di), "TestVerifyRegistration"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestVerifyRegistration(di *TestModuleDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := di.Consul.Client()
		catalogs, _, e := client.Catalog().Service("testservice", TestServiceID, (&api.QueryOptions{}).WithContext(ctx))
		g.Expect(e).To(Succeed(), "getting service catalog should not fail")
		g.Expect(catalogs).ToNot(BeEmpty(), "service catalog should not be empty")
		var svc *api.CatalogService
		for i := range catalogs {
			if TestServiceID == catalogs[i].ServiceID {
				svc = catalogs[i]
				break
			}
		}
		g.Expect(svc).ToNot(BeNil(), "service catalog should contain expected instance")
	}
}
