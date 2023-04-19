package actuator_tests

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/actuator_tests/testdata"
	actuatorinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/loggers"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/utils/gomega"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"strings"
	"testing"
)

var _ = log.New("Test")

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
			fx.Invoke(ConfigureSecurity),
		),
		test.GomegaSubTest(SubTestLoggersWithAccess(mockedSecurityAdmin()), "TestLoggersWithAccess"),
		test.GomegaSubTest(SubTestLoggersWithoutAccess(mockedSecurityNonAdmin()), "TestLoggersWithoutAccess"),
		test.GomegaSubTest(SubTestLoggersWithoutAuth(), "TestLoggersWithoutAuth"),
		test.GomegaSubTest(SubTestChangeLoggers(mockedSecurityAdmin()), "TestChangeLoggers"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLoggersWithAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)
		// with admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertLoggersV3Response(t, g, resp.Response)

		// with admin security GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/bootstrap", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertSingleLoggerV3Response(t, g, resp.Response)
	}
}

func SubTestLoggersWithoutAccess(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)

		// with non-admin security GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusForbidden)

		// with non-admin security GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/test", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusForbidden)
	}
}

func SubTestLoggersWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// regular GET
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers", nil, defaultRequestOptions())
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)

		// regular GET with name
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/test", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusUnauthorized)
	}
}

func SubTestChangeLoggers(secOpts sectest.SecurityContextOptions) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		ctx = sectest.ContextWithSecurity(ctx, secOpts)
		// send POST
		const newLevel = "ERROR"
		body := fmt.Sprintf(`{"configuredLevel":"%s"}`, newLevel)
		req := webtest.NewRequest(ctx, http.MethodPost, "/admin/loggers/test", strings.NewReader(body),
			webtest.ContentType("application/json"))
		resp := webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusNoContent)

		// check Result
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/loggers/test", nil, defaultRequestOptions())
		resp = webtest.MustExec(ctx, req)
		assertResponse(t, g, resp.Response, http.StatusOK, "Content-Type", actuator.ContentTypeSpringBootV3)
		assertSingleLoggerV3Response(t, g, resp.Response, func(body []byte) {
			g.Expect(body).To(HaveJsonPathWithValue("$.effectiveLevel", newLevel), "logger's effectiveLevel should be updated")
			g.Expect(body).To(HaveJsonPathWithValue("$.configuredLevel", newLevel), "logger's configuredLevel should be updated")
		})
	}
}

/*************************
	Common Helpers
 *************************/

func assertLoggersV3Response(_ *testing.T, g *gomega.WithT, resp *http.Response) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `loggers response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$.levels"), "loggers response should contains 'levels'")
	g.Expect(body).To(HaveJsonPath("$.loggers", ), "loggers response should contains 'loggers'")
	g.Expect(body).To(HaveJsonPath("$.loggers[*].effectiveLevel"), "loggers response should contains 'effectiveLevel'")
}

func assertSingleLoggerV3Response(_ *testing.T, g *gomega.WithT, resp *http.Response, extraChecks ...func(body []byte)) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `loggers response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$.effectiveLevel"), "loggers response should contains 'effectiveLevel'")
	for _, fn := range extraChecks {
		fn(body)
	}
}

