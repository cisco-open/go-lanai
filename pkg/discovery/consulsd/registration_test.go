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
	"fmt"
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

/*************************
	Tests
 *************************/

type TestRegDI struct {
	fx.In
	Consul              *consul.Connection
	AppContext          *bootstrap.ApplicationContext
	DiscoveryProperties consulsd.DiscoveryProperties
}

func TestRegistration(t *testing.T) {
	di := TestRegDI{}
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
		apptest.WithFxOptions(
			fx.Provide(consulsd.BindDiscoveryProperties),
		),

		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestNewRegistration(&di), "TestNewRegistration"),
		test.GomegaSubTest(SubTestRegisterWithProperties(&di), "TestRegisterWithProperties"),
		test.GomegaSubTest(SubTestRegisterWithCustomizers(&di), "TestRegisterWithCustomizers"),
		//test.GomegaSubTest(SubTestDeregister(&di), "TestDeregister"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestRegisterWithProperties(di *TestRegDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const serviceIdOverride = `testservice-8080-664d91a5ba`
		registrar := consulsd.NewServiceRegistrar(di.Consul)
		registration := consulsd.NewRegistration(ctx,
			consulsd.RegistrationWithProperties(&di.DiscoveryProperties),
			consulsd.RegistrationWithAppContext(di.AppContext),
		)
		AssertRegistration(g, registration, NewExpectedReg())

		ApplyServiceIDOverride(registration, serviceIdOverride)
		e := registrar.Register(ctx, registration)
		g.Expect(e).To(Succeed(), "register should not fail")
		VerifyConsul(ctx, g, di.Consul, NewExpectedReg(func(reg *ExpectedReg) {
			reg.ServiceID = serviceIdOverride
		}))

		e = registrar.Deregister(ctx, registration)
		g.Expect(e).To(Succeed(), "deregister should not fail")
		VerifyConsul(ctx, g, di.Consul, NewExpectedReg(func(reg *ExpectedReg) {
			reg.ServiceID = serviceIdOverride
			reg.Absent = true
		}))
	}
}

func SubTestRegisterWithCustomizers(di *TestRegDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const serviceIdOverride = `testservice-8080-d8755f792d`
		registrar := consulsd.NewServiceRegistrar(di.Consul)
		registration := consulsd.NewRegistration(ctx,
			consulsd.RegistrationWithProperties(&di.DiscoveryProperties),
			consulsd.RegistrationWithAppContext(di.AppContext),
			consulsd.RegistrationWithCustomizers(
				discovery.NewPropertiesBasedCustomizer(di.AppContext, nil),
				discovery.NewBuildInfoCustomizer(),
			),
		)
		AssertRegistration(g, registration, NewExpectedReg(WithComponentInfo(), WithBuildInfo()))

		ApplyServiceIDOverride(registration, serviceIdOverride)
		e := registrar.Register(ctx, registration)
		g.Expect(e).To(Succeed(), "register should not fail")
		VerifyConsul(ctx, g, di.Consul, NewExpectedReg(WithComponentInfo(), WithBuildInfo(), func(reg *ExpectedReg) {
			reg.ServiceID = serviceIdOverride
		}))

		e = registrar.Deregister(ctx, registration)
		g.Expect(e).To(Succeed(), "deregister should not fail")
		VerifyConsul(ctx, g, di.Consul, NewExpectedReg(func(reg *ExpectedReg) {
			reg.ServiceID = serviceIdOverride
			reg.Absent = true
		}))
	}
}

func SubTestNewRegistration(di *TestRegDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		reg := consulsd.NewRegistration(ctx,
			consulsd.RegistrationWithProperties(&di.DiscoveryProperties),
			consulsd.RegistrationWithAppContext(di.AppContext),
		).(*consulsd.ServiceRegistration)
		AssertRegistration(g, reg, NewExpectedReg())
		reg.SetID("override-id")
		g.Expect(reg.AgentServiceRegistration.ID).To(Equal("override-id"), `[%s] should be correct`, "ID")
		reg.SetName("override-name")
		g.Expect(reg.AgentServiceRegistration.Name).To(Equal("override-name"), `[%s] should be correct`, "Name")
		reg.SetAddress("0.0.0.0")
		g.Expect(reg.AgentServiceRegistration.Address).To(Equal("0.0.0.0"), `[%s] should be correct`, "Name")
		reg.SetPort(9999)
		g.Expect(reg.AgentServiceRegistration.Port).To(Equal(9999), `[%s] should be correct`, "Port")
		reg.AddTags("tag1", "tag2", "tag1", "tag1", "tag2")
		g.Expect(reg.AgentServiceRegistration.Tags).To(ContainElements("tag1", "tag2"), `[%s] should be correct`, "Tags")
		reg.RemoveTags("tag1", "tag2", "tag1", "tag2")
		g.Expect(reg.AgentServiceRegistration.Tags).ToNot(ContainElement("tag1"), `[%s] should be correct`, "Tags")
		g.Expect(reg.AgentServiceRegistration.Tags).ToNot(ContainElement("tag2"), `[%s] should be correct`, "Tags")
		reg.SetMeta("test", "value")
		g.Expect(reg.AgentServiceRegistration.Meta).To(HaveKeyWithValue("test", "value"), `[%s] should be correct`, "Meta")
		reg.SetMeta("test", nil)
		g.Expect(reg.AgentServiceRegistration.Meta).ToNot(HaveKeyWithValue("test", "value"), `[%s] should be correct`, "Meta")
	}
}

/*************************
	Helpers
 *************************/

type ExpectedReg struct {
	ServiceID               string
	Name                    string
	TagRegExps              []string
	Port                    int
	Address                 string
	HealthURL               string
	HealthInterval          string
	HealthDeregisterTimeout string
	Absent                  bool
}

func (r ExpectedReg) IDRegExp() string {
	return fmt.Sprintf(`%s-%d-[0-9a-f]+`, r.Name, r.Port)
}

func NewExpectedReg(opts ...func(reg *ExpectedReg)) *ExpectedReg {
	reg := ExpectedReg{
		Name: "testservice",
		TagRegExps: []string{
			"secure=false", "test1=true", "test2=true",
		},
		Port:                    8080,
		Address:                 "127.0.0.1",
		HealthURL:               "http://127.0.0.1:8080/test/admin/health",
		HealthInterval:          "1s",
		HealthDeregisterTimeout: "15s",
	}
	for _, fn := range opts {
		fn(&reg)
	}
	return &reg
}

func WithComponentInfo() func(reg *ExpectedReg) {
	return func(reg *ExpectedReg) {
		reg.TagRegExps = append(reg.TagRegExps,
			discovery.TagInstanceUUID+`=[0-9a-f\-]+`,
			discovery.TagServiceName+`=.+`,
			discovery.TagComponentAttributes+`=.+`,
		)
	}
}

func WithBuildInfo() func(reg *ExpectedReg) {
	return func(reg *ExpectedReg) {
		reg.TagRegExps = append(reg.TagRegExps,
			discovery.TagBuildDateTime+`=.+`,
		)
	}
}

func ApplyServiceIDOverride(reg discovery.ServiceRegistration, serviceID string) {
	reg.SetID(serviceID)
	reg.AddTags(serviceID)
}

func AssertRegistration(g *gomega.WithT, registration discovery.ServiceRegistration, expected *ExpectedReg) {
	g.Expect(registration).To(BeAssignableToTypeOf(&consulsd.ServiceRegistration{}), "registration should be correct type")
	reg := registration.(*consulsd.ServiceRegistration)
	g.Expect(reg).ToNot(BeNil(), "registration should not be nil")
	g.Expect(reg.Kind).To(Equal(api.ServiceKindTypical), "registration should have correct '%s'", "Kind")
	g.Expect(reg.ID()).To(MatchRegexp(expected.IDRegExp()), "registration should have correct '%s'", "ID")
	g.Expect(reg.Name()).To(Equal(expected.Name), "registration should have correct '%s'", "Name")
	g.Expect(reg.Port()).To(Equal(expected.Port), "registration should have correct '%s'", "Port")
	g.Expect(reg.Address()).To(Equal(expected.Address), "registration should have correct '%s'", "Address")
	g.Expect(reg.Check).ToNot(BeNil(), "registration should have non-nil '%s'", "Check")
	g.Expect(reg.Check.HTTP).To(Equal(expected.HealthURL), "registration should have correct '%s'", "Check.HTTP")
	g.Expect(reg.Check.Interval).To(Equal(expected.HealthInterval), "registration should have correct '%s'", "Check.Interval")
	g.Expect(reg.Check.DeregisterCriticalServiceAfter).To(Equal(expected.HealthDeregisterTimeout), "registration should have correct '%s'", "Check.DeregisterCriticalServiceAfter")
	for _, tag := range expected.TagRegExps {
		g.Expect(reg.Tags()).To(ContainElement(MatchRegexp(tag)), "registration tags should contain '%s'", tag)
	}
}

func AssertCatalogService(g *gomega.WithT, svc *api.CatalogService, expected *ExpectedReg) {
	g.Expect(svc).ToNot(BeNil(), "service catalog should not be nil")
	g.Expect(svc.ServiceID).To(MatchRegexp(expected.IDRegExp()), "service catalog should have correct '%s'", "ServiceID")
	g.Expect(svc.ServiceName).To(Equal(expected.Name), "service catalog should have correct '%s'", "ServiceName")
	g.Expect(svc.ServicePort).To(Equal(expected.Port), "service catalog should have correct '%s'", "ServicePort")
	g.Expect(svc.ServiceAddress).To(Equal(expected.Address), "service catalog should have correct '%s'", "Address")
	for _, tag := range expected.TagRegExps {
		g.Expect(svc.ServiceTags).To(ContainElement(MatchRegexp(tag)), "service catalog tags should contain '%s'", tag)
	}
}

func VerifyConsul(ctx context.Context, g *gomega.WithT, consulConn *consul.Connection, expected *ExpectedReg) {
	client := consulConn.Client()
	catalogs, _, e := client.Catalog().Service(expected.Name, expected.ServiceID, (&api.QueryOptions{}).WithContext(ctx))
	g.Expect(e).To(Succeed(), "getting service catalog should not fail")
	if len(catalogs) == 0 && expected.Absent {
		return
	}
	g.Expect(catalogs).ToNot(BeEmpty(), "service catalog should not be empty")
	var svc *api.CatalogService
	for i := range catalogs {
		if expected.ServiceID == catalogs[i].ServiceID {
			svc = catalogs[i]
			break
		}
	}
	if expected.Absent {
		g.Expect(svc).To(BeNil(), "service catalog should not contain expected instance")
	} else {
		g.Expect(svc).ToNot(BeNil(), "service catalog should contain expected instance")
		AssertCatalogService(g, svc, expected)
	}
}
