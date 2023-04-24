package actuatortest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/apilist"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/env"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	healthep "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health/endpoint"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/loggers"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/actuatortest/testdata"
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

/*************************
	Tests
 *************************/

func TestActuatorEndpoints(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(
			webtest.AddDefaultRequestOptions(webtest.Headers("Accept", "application/json")),
		),
		sectest.WithMockedMiddleware(),
		WithEndpoints(DisableAllEndpoints()),
		apptest.WithModules(
			health.Module, healthep.Module, loggers.Module, env.Module, apilist.Module,
		),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Provide(testdata.NewMockedHealthIndicator),
			fx.Invoke(ConfigureHealth),
		),
		apptest.WithDI(),
		test.GomegaSubTest(SubTestHealthWithDetails(), "TestHealthWithDetails"),
		test.GomegaSubTest(SubTestLoggersEndpoint(), "TestLoggersEndpoint"),
		test.GomegaSubTest(SubTestEnvEndpoint(), "TestEnvEndpoint"),
		test.GomegaSubTest(SubTestAPIListEndpoint(), "TestAPIListEndpoint"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestHealthWithDetails() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
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

func SubTestLoggersEndpoint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertLoggersResponse(t, resp.Response)

		// with admin security GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/bootstrap", nil)
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertLoggersResponse(t, resp.Response, ExpectLoggersSingleEntry())
	}
}

func SubTestAPIListEndpoint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/apilist", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertAPIListResponse(t, resp.Response)
	}
}

func SubTestEnvEndpoint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		AssertEnvResponse(t, resp.Response)
	}
}

/*************************
	Helpers
 *************************/

func v2RequestOptions() webtest.RequestOptions {
	return func(req *http.Request) {
		req.Header.Set("Accept", actuator.ContentTypeSpringBootV2)
	}
}

func assertResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedStatus int, expectedHeaders ...string) {
	g.Expect(resp).ToNot(BeNil(), "endpoint should have response")
	g.Expect(resp.StatusCode).To(BeEquivalentTo(expectedStatus))
	for i := range expectedHeaders {
		if i%2 == 1 || i+1 >= len(expectedHeaders) {
			continue
		}
		k := expectedHeaders[i]
		v := expectedHeaders[i+1]
		g.Expect(resp.Header.Get(k)).To(BeEquivalentTo(v), "response header should contains [%s]='%s'", k, v)
	}
}