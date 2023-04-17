package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	actuatorinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/loggers"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Tests
 *************************/

func TestLoggersEndpoint(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(),
		apptest.WithModules(
			actuatorinit.Module, actuator.Module, access.Module, errorhandling.Module,
			loggers.Module,
		),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Provide(testdata.NewMockedHealthIndicator),
			fx.Invoke(ConfigureSecurity),
		),
		test.GomegaSubTest(SubTestLoggersWithPermissions(), "TestLoggersWithPermissions"),
		test.GomegaSubTest(SubTestLoggersWithoutPermissions(), "TestLoggersWithoutPermissions"),
		test.GomegaSubTest(SubTestLoggersWithoutAuth(), "TestLoggersWithoutAuth"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLoggersWithPermissions() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, mockedSecurityAdmin())
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		// TODO verify response body

		// with admin security GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/actuator", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		// TODO verify response body
	}
}

func SubTestLoggersWithoutPermissions() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, mockedSecurityNonAdmin())

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusForbidden)
		// TODO verify response body

		// with non-admin security GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/actuator", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		// TODO verify response body
	}
}

func SubTestLoggersWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)
		// TODO verify response body

		// regular GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/actuator", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)
		// TODO verify response body
	}
}

