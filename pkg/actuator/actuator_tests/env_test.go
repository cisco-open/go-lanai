package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/env"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/actuatortest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

/*************************
	Tests
 *************************/

func TestEnvEndpoint(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(webtest.AddDefaultRequestOptions(v3RequestOptions())),
		sectest.WithMockedMiddleware(),
		actuatortest.WithEndpoints(actuatortest.DisableAllEndpoints()),
		apptest.WithModules(env.Module),
		apptest.WithConfigFS(testdata.TestConfigFS),
		test.GomegaSubTest(SubTestEnvWithAccess(mockedSecurityAdmin()), "TestEnvWithAccess"),
		test.GomegaSubTest(SubTestEnvWithoutAccess(mockedSecurityNonAdmin()), "TestEnvWithoutAccess"),
		test.GomegaSubTest(SubTestEnvWithoutAuth(), "TestEnvWithoutAuth"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestEnvWithAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		actuatortest.AssertEnvResponse(t, resp.Response)
	}
}

func SubTestEnvWithoutAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusForbidden)
	}
}

func SubTestEnvWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)
	}
}

/*************************
	Common Helpers
 *************************/


