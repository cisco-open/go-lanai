package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health/endpoint"
	actuatorinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/utils/gomega"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
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
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(),
		apptest.WithModules(
			actuatorinit.Module, actuator.Module, access.Module, errorhandling.Module,
			health.Module, healthep.Module,
		),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Provide(testdata.NewMockedHealthIndicator),
			fx.Invoke(ConfigureSecurity),
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
		apptest.WithModules(
			actuatorinit.Module, actuator.Module, access.Module, errorhandling.Module,
			health.Module, healthep.Module,
		),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithProperties(
			"management.endpoint.health.show-details: custom",
			"management.endpoint.health.show-components: custom",
		),
		apptest.WithFxOptions(
			fx.Provide(testdata.NewMockedHealthIndicator),
			fx.Invoke(ConfigureSecurity),
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
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertHealthV3Response(t, g, resp.Response, health.StatusUp, true, true)

		// with admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		assertHealthV2Response(t, g, resp.Response, health.StatusUp, true, true)
	}
}

func SubTestHealthWithoutDetails(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertHealthV3Response(t, g, resp.Response, health.StatusUp, false, false)

		// with non-admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		assertHealthV2Response(t, g, resp.Response, health.StatusUp, false, false)
	}
}

func SubTestHealthWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertHealthV3Response(t, g, resp.Response, health.StatusUp, false, false)

		// with admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		assertHealthV2Response(t, g, resp.Response, health.StatusUp, false, false)
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
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusServiceUnavailable, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertHealthV3Response(t, g, resp.Response, health.StatusDown, true, true)

		// out of service
		di.MockedIndicator.Status = health.StatusOutOfService
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusServiceUnavailable, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertHealthV2Response(t, g, resp.Response, health.StatusOutOfService, true, true)
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
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusServiceUnavailable, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertHealthV3Response(t, g, resp.Response, health.StatusDown, false, false)

		// out of service
		di.MockedIndicator.Status = health.StatusOutOfService
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusServiceUnavailable, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertHealthV2Response(t, g, resp.Response, health.StatusOutOfService, false, false)
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

func assertHealthV3Response(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedStatus health.Status, expectDetails, expectComponents bool) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `health response body should be readable`)
	g.Expect(body).To(HaveJsonPathWithValue("$.status", expectedStatus.String()), "health response should have status [%v]", expectedStatus)

	if expectComponents {
		g.Expect(body).To(HaveJsonPath("$..components"), "v3 health response should have components")
	} else {
		g.Expect(body).NotTo(HaveJsonPath("$..components"), "v3 health response should not have components")
	}

	if expectDetails {
		g.Expect(body).To(HaveJsonPath("$..details"), "v3 health response should have details")
	} else {
		g.Expect(body).NotTo(HaveJsonPath("$..details"), "v3 health response should not have details")
	}
}

func assertHealthV2Response(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedStatus health.Status, expectDetails, expectComponents bool) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `health response body should be readable`)
	g.Expect(body).To(HaveJsonPathWithValue("$.status", ContainElement(expectedStatus.String())), "health response should have status [%v]", expectedStatus)

	if expectComponents {
		g.Expect(body).To(HaveJsonPath("$..details"), "v2 health response should have components")
	} else {
		g.Expect(body).NotTo(HaveJsonPath("$..details"), "v2 health response should not have components")
	}

	if expectDetails {
		g.Expect(body).To(HaveJsonPath("$..detailed"), "v2 health response should have details")
	} else {
		g.Expect(body).NotTo(HaveJsonPath("$..detailed"), "v2 health response should not have details")
	}
}

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
