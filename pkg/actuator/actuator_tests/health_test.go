package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health/endpoint"
	actuatorinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
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
		test.GomegaSubTest(SubTestHealthWithPermissions(), "TestHealthWithPermissions"),
		test.GomegaSubTest(SubTestHealthWithoutPermissions(), "TestHealthWithoutPermissions"),
		test.GomegaSubTest(SubTestHealthWithoutAuth(), "TestHealthWithoutAuth"),
		test.GomegaSubTest(SubTestHealthNegativeCases(di), "TestHealthNegativeCases"),
		test.GomegaSubTest(SubTestHealthIndicator(di), "TestHealthIndicator"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestHealthWithPermissions() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, mockedSecurityAdmin())

		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		// TODO verify response body

		// with admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		// TODO verify response body
	}
}

func SubTestHealthWithoutPermissions() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, mockedSecurityNonAdmin())

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		// TODO verify response body

		// with non-admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
	}
}

func SubTestHealthWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		// TODO verify response body

		// with admin security GET V2
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, v2RequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
	}
}

func SubTestHealthNegativeCases(di *HealthTestDI) test.GomegaSubTestFunc {
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
		// TODO verify response body

		// out of service
		di.MockedIndicator.Status = health.StatusOutOfService
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/health", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusServiceUnavailable, "Content-Type", actuator.ContentTypeSpringBootV3)
		// TODO verify response body
	}
}

/*************************
	Common Helpers
 *************************/

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
