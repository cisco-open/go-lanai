package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/alive"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/info"
	actuatorinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
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
	Tests
 *************************/

// TestSimpleAdminEndpoints test simple endpoints like info, alive
func TestSimpleAdminEndpoints(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(),
		apptest.WithModules(
			actuatorinit.Module, actuator.Module, access.Module, errorhandling.Module,
			info.Module, alive.Module,
		),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Invoke(ConfigureSecurity),
		),
		test.GomegaSubTest(SubTestInfoEndpointV3(), "TestInfoEndpointV3"),
		test.GomegaSubTest(SubTestInfoEndpointV2(), "TestInfoEndpointV2"),
		test.GomegaSubTest(SubTestAliveEndpoint(), "TestAliveEndpoint"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestInfoEndpointV3() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/info", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertInfoResponse(t, g, resp.Response)

		// By name, currently no supported
		//req = webtest.NewRequest(ctx, http.MethodGet, "/admin/info/app", nil, defaultRequestOptions())
		//resp = webtest.MustExec(ctx, req)
		//assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
	}
}

func SubTestInfoEndpointV2() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/info", nil, v2RequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV2)
		assertInfoResponse(t, g, resp.Response)

		// By name, currently no supported
		//req = webtest.NewRequest(ctx, http.MethodGet, "/admin/info/app", nil, defaultRequestOptions())
		//resp = webtest.MustExec(ctx, req)
		//assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
	}
}

func SubTestAliveEndpoint() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/alive", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
	}
}

/*************************
	Common Helpers
 *************************/

func assertInfoResponse(_ *testing.T, g *gomega.WithT, resp *http.Response) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `info response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$.build-info.version"), "info response should build version")
}