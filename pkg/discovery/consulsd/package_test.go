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

package consulsd_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/discovery/consulsd"
	"github.com/cisco-open/go-lanai/pkg/discovery/consulsd/testdata"
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
const TestRegisterFuzzyJsonPathMeta = `$.Meta`
const TestServiceID = `testservice-8080-d8755f792d`

/*************************
	Test Setup
 *************************/

func NewTestServiceIDOverrider() discovery.ServiceRegistrationCustomizer {
	return discovery.ServiceRegistrationCustomizerFunc(func(ctx context.Context, reg discovery.ServiceRegistration) {
		reg.SetID(TestServiceID)
		reg.AddTags(TestServiceID)
	})
}

/*************************
	Tests
 *************************/

type TestModuleDI struct {
	fx.In
	Consul              *consul.Connection
	AppContext          *bootstrap.ApplicationContext
	DiscoveryProperties consulsd.DiscoveryProperties
	DiscoveryClient     discovery.Client
	Registration        discovery.ServiceRegistration
	Registrar           discovery.ServiceRegistrar
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
				TestRegisterFuzzyJsonPathMeta,
			))),
		),
		apptest.WithBootstrapConfigFS(testdata.TestBootstrapFS),
		apptest.WithConfigFS(testdata.TestApplicationFS),
		apptest.WithModules(consulsd.Module),
		apptest.WithFxOptions(
			fx.Provide(fx.Annotate(NewTestServiceIDOverrider, fx.ResultTags(`group:"discovery"`))),
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
