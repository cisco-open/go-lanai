package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/env"
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

func TestEnvEndpoint(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(),
		apptest.WithModules(
			actuatorinit.Module, actuator.Module, access.Module, errorhandling.Module,
			env.Module,
		),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Invoke(ConfigureSecurity),
		),
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
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertEnvV3Response(t, g, resp.Response)
	}
}

func SubTestEnvWithoutAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusForbidden)
	}
}

func SubTestEnvWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)
	}
}

/*************************
	Common Helpers
 *************************/

func assertEnvV3Response(_ *testing.T, g *gomega.WithT, resp *http.Response) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `env response body should be readable`)
	g.Expect(body).To(HaveJsonPathWithValue("$.activeProfiles[0]", "test"), "env response should contains correct active profiles")
	g.Expect(body).To(HaveJsonPath("$.propertySources"), "env response should contains propertySources")
	g.Expect(body).To(HaveJsonPath("$.propertySources[0]"), "env response should contains non-empty propertySources")
}
