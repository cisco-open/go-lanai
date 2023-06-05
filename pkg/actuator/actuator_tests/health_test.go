package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health/endpoint"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/actuatortest"
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/actuatortest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Test Setup
 *************************/

func ConfigureHealth(healthReg health.Registrar, mock *testdata.MockedHealthIndicator) {
	healthReg.MustRegister(mock)
}

func ConfigureCustomHealthDisclosure(healthReg health.Registrar) {
	healthReg.MustRegister(health.DisclosureControlFunc(func(ctx context.Context) (ok bool) {
		ok, _ = tokenauth.ScopesApproved(SpecialScopeAdmin)(security.Get(ctx))
		return
	}))
}

type HealthTestDI struct {
	fx.In
	TestDI
	HealthIndicator health.Indicator
	MockedIndicator *testdata.MockedHealthIndicator
}

/*************************
	Tests
 *************************/

func TestHealthEndpoint(t *testing.T) {
	di := &HealthTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(webtest.AddDefaultRequestOptions(v3RequestOptions())),
		sectest.WithMockedMiddleware(),
		actuatortest.WithEndpoints(actuatortest.DisableAllEndpoints()),
		apptest.WithModules(health.Module, healthep.Module),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Provide(testdata.NewMockedHealthIndicator),
			fx.Invoke(ConfigureHealth),
		),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestHealthWithDetails(mockedSecurityAdmin()), "TestHealthWithDetails"),
		test.GomegaSubTest(SubTestHealthWithoutDetails(mockedSecurityNonAdmin()), "TestHealthWithoutDetails"),
		test.GomegaSubTest(SubTestHealthWithoutAuth(), "TestHealthWithoutAuth"),
		test.GomegaSubTest(SubTestHealthDownWithDetails(di, mockedSecurityAdmin()), "TestHealthDownWithDetails"),
		test.GomegaSubTest(SubTestHealthDownWithoutDetails(di), "TestHealthDownWithoutDetails"),
		test.GomegaSubTest(SubTestHealthIndicator(di), "TestHealthIndicator"),
	)
}

func TestHealthWithCustomDisclosure(t *testing.T) {
	di := &HealthTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(),
		actuatortest.WithEndpoints(actuatortest.DisableAllEndpoints()),
		apptest.WithModules(health.Module, healthep.Module),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithProperties(
			"management.endpoint.health.show-details: custom",
			"management.endpoint.health.show-components: custom",
		),
		apptest.WithFxOptions(
			fx.Provide(testdata.NewMockedHealthIndicator),
			fx.Invoke(ConfigureHealth),
			fx.Invoke(ConfigureCustomHealthDisclosure),
		),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestHealthWithDetails(mockedSecurityScopedAdmin()), "TestHealthWithPermissions"),
		test.GomegaSubTest(SubTestHealthWithoutDetails(mockedSecurityAdmin()), "TestHealthWithoutPermissions"),
		test.GomegaSubTest(SubTestHealthWithoutAuth(), "TestHealthWithoutAuth"),
		test.GomegaSubTest(SubTestHealthDownWithDetails(di, mockedSecurityScopedAdmin()), "TestHealthDownWithDetails"),
		test.GomegaSubTest(SubTestHealthDownWithoutDetails(di), "TestHealthDownWithoutDetails"),
		test.GomegaSubTest(SubTestHealthIndicator(di), "TestHealthIndicator"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestHealthWithDetails(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)

		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertHealthResponse(t, resp.Response, ExpectHealthDetails(), ExpectHealthComponents("test"))

		// with admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		AssertHealthResponse(t, resp.Response, ExpectHealthDetails(), ExpectHealthComponents("test"))
	}
}

func SubTestHealthWithoutDetails(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertHealthResponse(t, resp.Response)

		// with non-admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		AssertHealthResponse(t, resp.Response)
	}
}

func SubTestHealthWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertHealthResponse(t, resp.Response)

		// with admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		AssertHealthResponse(t, resp.Response)
	}
}

func SubTestHealthDownWithDetails(di *HealthTestDI, secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)
		// negative cases
		defer func() {
			di.MockedIndicator.Status = health.StatusUp
		}()
		// down
		di.MockedIndicator.Status = health.StatusDown
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusServiceUnavailable, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertHealthResponse(t, resp.Response, ExpectHealth(health.StatusDown), ExpectHealthDetails(), ExpectHealthComponents("test"))

		// out of service
		di.MockedIndicator.Status = health.StatusOutOfService
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil)
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusServiceUnavailable, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertHealthResponse(t, resp.Response, ExpectHealth(health.StatusOutOfService), ExpectHealthDetails(), ExpectHealthComponents("test"))
	}
}

func SubTestHealthDownWithoutDetails(di *HealthTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// negative cases
		defer func() {
			di.MockedIndicator.Status = health.StatusUp
		}()
		// down
		di.MockedIndicator.Status = health.StatusDown
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusServiceUnavailable, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertHealthResponse(t, resp.Response, ExpectHealth(health.StatusDown))

		// out of service
		di.MockedIndicator.Status = health.StatusOutOfService
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil)
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusServiceUnavailable, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertHealthResponse(t, resp.Response, ExpectHealth(health.StatusOutOfService), )
	}
}


func SubTestHealthIndicator(di *HealthTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		opts := health.Options{
			ShowDetails:    true,
			ShowComponents: true,
		}
		// show everything
		h := di.HealthIndicator.Health(ctx, opts)
		assertHealth(t, g, h, health.StatusUp, opts)

		// show components but no details
		opts.ShowDetails = false
		h = di.HealthIndicator.Health(ctx, opts)
		assertHealth(t, g, h, health.StatusUp, opts)

		// no components
		opts.ShowComponents = false
		h = di.HealthIndicator.Health(ctx, opts)
		assertHealth(t, g, h, health.StatusUp, opts)

		// down
		di.MockedIndicator.Status = health.StatusDown
		defer func() {
			di.MockedIndicator.Status = health.StatusUp
		}()
		h = di.HealthIndicator.Health(ctx, opts)
		assertHealth(t, g, h, health.StatusDown, opts)
	}
}

/*************************
	Common Helpers
 *************************/

func assertHealth(t *testing.T, g *gomega.WithT, h health.Health, expected health.Status, expectedOpts health.Options) {
	g.Expect(h).To(Not(BeNil()), `Health status should not be nil`)
	g.Expect(h.Status()).To(BeEquivalentTo(expected), `Health [%s] status should be %v`, h.Description(), expected)

	switch v := h.(type) {
	case *health.DetailedHealth:
		// check details
		if expectedOpts.ShowDetails {
			g.Expect(v.Details).ToNot(BeEmpty(), "Detailed health should contains details")
		} else {
			g.Expect(v.Details).To(BeEmpty(), "Detailed health should not contains details")
		}
	case *health.CompositeHealth:
		// check components
		if expectedOpts.ShowComponents {
			g.Expect(v.Components).ToNot(BeEmpty(), "Composite health should contains components")
		} else {
			g.Expect(v.Components).To(BeEmpty(), "Composite health should not contains components")
		}
		// recursively assert components
		for _, comp := range v.Components {
			if comp.Description() == "ping" {
				continue
			}
			assertHealth(t, g, comp, expected, expectedOpts)
		}
	}
}