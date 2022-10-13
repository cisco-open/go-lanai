package examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"embed"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Examples
 *************************/

const (
	TestHeader        = "X-MOCK-AUTH"
)

// TestUseDefaultSecurityScopeMocking
// apptest.Bootstrap and sectest.WithMockedScopes are required for usage of scope package
func TestUseDefaultSecurityScopeMocking(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		sectest.WithMockedScopes(),
		test.GomegaSubTest(SubTestExampleUseScope(), "UseScope"),
	)
	// Any sub tests can use "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope" as usual
}

//go:embed example-custom-scopes.yml
var customMockingConfigFS embed.FS

// TestUseCustomSecurityScopeMocking
// apptest.Bootstrap and sectest.WithMockedScopes are required for usage of scope package
func TestUseCustomSecurityScopeMocking(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		sectest.WithMockedScopes(customMockingConfigFS), // use custom config as embeded configuration
		test.GomegaSubTest(SubTestExampleUseScope(), "UseScope"),
	)
	// Any sub tests can use "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope" as usual
}

// TestCurrentSecurityContextMocking
// apptest.Bootstrap and sectest.WithMockedScopes are NOT required for usage of sectest.WithMockedSecurity
func TestCurrentSecurityContextMocking(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestExampleMockCurrentSecurity(), "MockCurrentSecurity"),
	)
}

// TestMockBothCurrentSecurityAndScope
// apptest.Bootstrap and sectest.WithMockedScopes are NOT required for usage of sectest.WithMockedSecurity
func TestMockBothCurrentSecurityAndScope(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestExampleMockBoth(), "MockBoth"),
	)
}

type testDI struct {
	fx.In
	AuthMocker sectest.MWMocker
}

func TestWithMockedServer(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(),
		apptest.WithModules(basicauth.Module, access.Module, errorhandling.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Invoke(registerTestController, registerTestSecurity),
		),
		test.GomegaSubTest(SubTestExampleWithMockedServer(di), "MockedServerWithMockedContext"),
	)
}

func TestWithRealServer(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		sectest.WithMockedMiddleware(
			// Custom Mocker is required for real server
			sectest.MWCustomMocker(sectest.MWMockFunc(realServerMockFunc)),
			//MWCustomConfigurer(security.ConfigurerFunc(realServerSecConfigurer)),
		),
		apptest.WithModules(basicauth.Module, access.Module, errorhandling.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Invoke(registerTestController, registerTestSecurity),
		),
		test.GomegaSubTest(SubTestExampleWithRealServer(di), "RealServerWithHeader"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestExampleUseScope() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		toTest := &TestTarget{}
		e := toTest.DoSomethingWithinSecurityScope(ctx)
		g.Expect(e).To(Succeed(), "scoped operation shouldn't returns error")
	}
}

func SubTestExampleMockCurrentSecurity() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.WithMockedSecurity(ctx, func(d *sectest.SecurityDetailsMock) {
			d.Username = "any-username"
			d.UserId = "any-user-id"
			d.TenantId = "any-tenant-id"
			d.TenantExternalId = "any-tenant-external-id"
			d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
			// see sectest.SecurityDetailsMock for more options
		})

		toTest := &TestTarget{}
		e := toTest.DoSomethingRequiringSecurity(ctx)
		g.Expect(e).To(Succeed(), "methods requiring security shouldn't returns error")
	}
}

func SubTestExampleMockBoth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// combined usage
		ctx = sectest.WithMockedSecurity(ctx, func(d *sectest.SecurityDetailsMock) {
			d.Username = "any-username"
			d.UserId = "any-user-id"
			d.TenantId = "any-tenant-id"
			d.TenantExternalId = "any-tenant-external-id"
			d.Permissions = utils.NewStringSet(security.SpecialPermissionSwitchTenant)
			// see sectest.SecurityDetailsMock for more options
		})

		toTest := &TestTarget{}
		e := toTest.DoSomethingWithinSecurityScope(ctx)
		g.Expect(e).To(Succeed(), "security-aware methods shouldn't returns error")
	}
}

func SubTestExampleWithMockedServer(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = sectest.WithMockedSecurity(ctx, securityMock())
		var req *http.Request
		var resp *http.Response
		req = webtest.NewRequest(ctx, http.MethodGet, TestSecuredURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp).To(Not(BeNil()), "response shouldn't be nil")
		g.Expect(resp.StatusCode).To(BeNumerically("==", 200), "response should be 200")
	}
}

func SubTestExampleWithRealServer(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = webtest.NewRequest(ctx, http.MethodGet, TestSecuredURL, nil)
		req.Header.Set(TestHeader, "yes")
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp).To(Not(BeNil()), "response shouldn't be nil")
		g.Expect(resp.StatusCode).To(BeNumerically("==", 200), "response should be 200")
	}
}

/*************************
	Helpers
 *************************/

func securityMock() sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		d.Username = "test-user"
		d.UserId = uuid.New().String()
		d.TenantId = uuid.New().String()
		d.TenantExternalId = "test-tenant"
		d.Permissions = utils.NewStringSet("TEST_PERMISSION")
	}
}

func realServerMockFunc(mc sectest.MWMockContext) security.Authentication {
	if mc.Request.Header.Get(TestHeader) == "" {
		return nil
	}
	return mockedAuth{}
}

//goland:noinspection GoUnusedFunction
func realServerSecConfigurer(ws security.WebSecurity) {
	ws = ws.Route(matcher.AnyRoute()).
		With(sectest.NewMockedMW().
			Mocker(sectest.MWMockFunc(realServerMockFunc)),
		)
}

type mockedAuth struct{}

func (a mockedAuth) Principal() interface{} {
	return "mocked"
}

func (a mockedAuth) Permissions() security.Permissions {
	return security.Permissions{}
}

func (a mockedAuth) State() security.AuthenticationState {
	return security.StateAuthenticated
}

func (a mockedAuth) Details() interface{} {
	return map[string]interface{}{}
}
