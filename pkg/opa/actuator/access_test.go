package opaactuator_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/alive"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/env"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/info"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opaactuator "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/actuator/testdata"
	opatest "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/actuatortest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/open-policy-agent/opa/sdk"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Test Setup
 *************************/

type secDI struct {
	fx.In
	SecReg         security.Registrar
	ActrReg        *actuator.Registrar
	ActrProperties actuator.ManagementProperties
	OPA            *sdk.OPA
}

func ConfigureSecurity(di secDI) {
	acCustomizer := opaactuator.NewAccessControlWithOPA(di.ActrProperties.Security, opa.RequestQueryWithPolicy("actuator/allow_endpoint"))
	di.ActrReg.MustRegister(acCustomizer)
}

/*************************
	Tests
 *************************/

func TestEndpointSecurity(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(webtest.AddDefaultRequestOptions(v3RequestOptions())),
		sectest.WithMockedMiddleware(),
		actuatortest.WithEndpoints(actuatortest.DisableAllEndpoints()),
		opatest.WithBundles(opatest.DefaultBundleFS, testdata.ActuatorBundleFS),
		apptest.WithModules(info.Module, env.Module, alive.Module),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Invoke(ConfigureSecurity),
		),
		test.GomegaSubTest(SubTestPublicAdminEndpoints(), "TestPublicAdminEndpoints"),
		test.GomegaSubTest(SubTestProtectedAdminEndpoints(), "TestProtectedAdminEndpoints"),
	)
}

/*************************
	Sub Tests
 *************************/
func SubTestPublicAdminEndpoints() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// info is controlled by OPA
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/info", nil)
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")

		// alive's security is disabled
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/alive", nil)
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")
	}
}

func SubTestProtectedAdminEndpoints() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// no auth
		req := webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusForbidden), "protected endpoint should reject requests without security")

		// non-admin
		ctx = sectest.ContextWithSecurity(ctx, mockedSecurityNonAdmin())
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusForbidden), "protected endpoint should reject requests with non admin client")

		// admin
		ctx = sectest.ContextWithSecurity(ctx, mockedSecurityScopedAdmin())
		req = webtest.NewRequest(ctx, http.MethodGet, "/admin/env", nil)
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "protected endpoint should accept requests with admin client")
	}
}

/*************************
	Common Helpers
 *************************/
