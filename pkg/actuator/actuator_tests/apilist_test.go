package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/apilist"
	actuatorinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/utils"
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

func TestAPIListEndpoint(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(),
		apptest.WithModules(
			actuatorinit.Module, actuator.Module, access.Module, errorhandling.Module,
			apilist.Module,
		),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithProperties(
			"management.endpoint.apilist.enabled: true",
			"management.endpoint.apilist.static-path: testdata/test-api-list.json",
		),
		apptest.WithFxOptions(
			fx.Invoke(ConfigureSecurity),
		),
		test.GomegaSubTest(SubTestAPIListWithAccess(mockedSecurityAdmin()), "TestAPIListWithAccess"),
		test.GomegaSubTest(SubTestAPIListWithoutAccess(mockedSecurityNonAdmin()), "TestAPIListWithoutAccess"),
		test.GomegaSubTest(SubTestAPIListWithoutAuth(), "TestAPIListWithoutAuth"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestAPIListWithAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/apilist", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertAPIListResponse(t, g, resp.Response)
	}
}

func SubTestAPIListWithoutAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/apilist", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusForbidden)
	}
}

func SubTestAPIListWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/apilist", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)
	}
}

/*************************
	Common Helpers
 *************************/

func assertAPIListResponse(_ *testing.T, g *gomega.WithT, resp *http.Response) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `apilist response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$..endpoint"), "apilist response should contain some endpoint field")
}
