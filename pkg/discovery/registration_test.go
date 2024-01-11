package discovery_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/consultest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

const TestRegisterFuzzyJsonPathTags = `$.Tags`
const TestRegisterFuzzyJsonPathMeta = `$.Meta`

/*************************
	Tests
 *************************/

type TestRegDI struct {
	fx.In
	Consul              *consul.Connection
	AppContext          *bootstrap.ApplicationContext
	DiscoveryProperties discovery.DiscoveryProperties
	Customizers         *discovery.Customizers
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
			fx.Provide(discovery.BindDiscoveryProperties),
			fx.Provide(discovery.NewCustomizers),
		),

		apptest.WithDI(&di),
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
		registration := discovery.NewRegistration(discovery.RegistrationWithProperties(di.AppContext, di.DiscoveryProperties))
		AssertRegistration(g, registration, NewExpectedReg())

		ApplyServiceIDOverride(registration, serviceIdOverride)
		e := discovery.Register(ctx, di.Consul, registration)
		g.Expect(e).To(Succeed(), "register should not fail")
		VerifyConsul(ctx, g, di.Consul, NewExpectedReg(func(reg *ExpectedReg) {
			reg.ServiceID = serviceIdOverride
		}))

		e = discovery.Deregister(ctx, di.Consul, registration)
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
		registration := discovery.NewRegistration(discovery.RegistrationWithProperties(di.AppContext, di.DiscoveryProperties))
		di.Customizers.Apply(ctx, registration)
		AssertRegistration(g, registration, NewExpectedReg(WithComponentInfo(), WithBuildInfo()))

		ApplyServiceIDOverride(registration, serviceIdOverride)
		e := discovery.Register(ctx, di.Consul, registration)
		g.Expect(e).To(Succeed(), "register should not fail")
		VerifyConsul(ctx, g, di.Consul, NewExpectedReg(WithComponentInfo(), WithBuildInfo(), func(reg *ExpectedReg) {
			reg.ServiceID = serviceIdOverride
		}))

		e = discovery.Deregister(ctx, di.Consul, registration)
		g.Expect(e).To(Succeed(), "deregister should not fail")
		VerifyConsul(ctx, g, di.Consul, NewExpectedReg(func(reg *ExpectedReg) {
			reg.ServiceID = serviceIdOverride
			reg.Absent = true
		}))
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
			"secure=false", "contextPath=/test", "test1=true", "test2=true",
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
			discovery.TAG_MANAGED_SERVICE,
			discovery.TAG_INSTANCE_ID+`=[0-9a-f\-]+`,
			discovery.TAG_SERVICE_NAME+`=.+`,
			discovery.TAG_COMPONENT_ATTRIBUTES+`=.+`,
		)
	}
}

func WithBuildInfo() func(reg *ExpectedReg) {
	return func(reg *ExpectedReg) {
		reg.TagRegExps = append(reg.TagRegExps,
			discovery.TAG_BUILD_DATE_TIME+`=.+`,
		)
	}
}

func ApplyServiceIDOverride(reg *api.AgentServiceRegistration, serviceID string) {
	reg.ID = serviceID
	reg.Tags = append(reg.Tags, serviceID)
}

func AssertRegistration(g *gomega.WithT, reg *api.AgentServiceRegistration, expected *ExpectedReg) {
	g.Expect(reg).ToNot(BeNil(), "registration should not be nil")
	g.Expect(reg.Kind).To(Equal(api.ServiceKindTypical), "registration should have correct '%s'", "Kind")
	g.Expect(reg.ID).To(MatchRegexp(expected.IDRegExp()), "registration should have correct '%s'", "ID")
	g.Expect(reg.Name).To(Equal(expected.Name), "registration should have correct '%s'", "Name")
	g.Expect(reg.Port).To(Equal(expected.Port), "registration should have correct '%s'", "Port")
	g.Expect(reg.Address).To(Equal(expected.Address), "registration should have correct '%s'", "Address")
	g.Expect(reg.Check).ToNot(BeNil(), "registration should have non-nil '%s'", "Check")
	g.Expect(reg.Check.HTTP).To(Equal(expected.HealthURL), "registration should have correct '%s'", "Check.HTTP")
	g.Expect(reg.Check.Interval).To(Equal(expected.HealthInterval), "registration should have correct '%s'", "Check.Interval")
	g.Expect(reg.Check.DeregisterCriticalServiceAfter).To(Equal(expected.HealthDeregisterTimeout), "registration should have correct '%s'", "Check.DeregisterCriticalServiceAfter")
	for _, tag := range expected.TagRegExps {
		g.Expect(reg.Tags).To(ContainElement(MatchRegexp(tag)), "registration tags should contain '%s'", tag)
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
