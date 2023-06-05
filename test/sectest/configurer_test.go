package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Setup
 *************************/

const (
	TestSecuredURL    = "/api/v1/secured"
	TestEntryPointURL = "/login"
	TestHeader        = "X-MOCK-AUTH"
)

type TestController struct{}

func registerTestController(reg *web.Registrar) {
	reg.MustRegister(&TestController{})
}

func (c *TestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("secured-get").Get(TestSecuredURL).
			EndpointFunc(c.Secured).Build(),
		rest.New("secured-post").Post(TestSecuredURL).
			EndpointFunc(c.Secured).Build(),
	}
}

func (c *TestController) Secured(_ context.Context, _ *http.Request) (interface{}, error) {
	return map[string]interface{}{
		"Message": "Yes",
	}, nil
}

type TestSecConfigurer struct{}

func (c *TestSecConfigurer) Configure(ws security.WebSecurity) {
	ws.Route(matcher.RouteWithPattern("/api/**")).
		With(
			basicauth.New().EntryPoint(redirect.NewRedirectWithRelativePath(TestEntryPointURL, false)),
		).
		With(access.New().Request(matcher.AnyRequest()).Authenticated()).
		With(errorhandling.New())
}

func registerTestSecurity(registrar security.Registrar) {
	cfg := TestSecConfigurer{}
	registrar.Register(&cfg)
}

/*************************
	Test
 *************************/

type testDI struct {
	fx.In
	AuthMocker MWMocker
}

func TestDefaultMWMocking(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		WithMockedMiddleware(),
		apptest.WithModules(basicauth.Module, access.Module, errorhandling.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Invoke(registerTestController, registerTestSecurity),
		),
		test.GomegaSubTest(SubTestWithoutMockedContext(di), "TestWithoutMockedContext"),
		test.GomegaSubTest(SubTestWithMockedContext(di), "TestWithMockedContext"),
	)
}

func TestCustomMWMocking(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		WithMockedMiddleware(
			MWRoute(matcher.RouteWithPattern("/api/**")),
			MWCondition(matcher.RequestWithMethods(http.MethodGet)),
			MWCustomMocker(nil), // enable autowired mode
			MWForceOverride(),
		),
		apptest.WithModules(basicauth.Module, access.Module, errorhandling.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(newCustomMocker),
			fx.Invoke(registerTestController, registerTestSecurity),
		),
		test.GomegaSubTest(SubTestCustomOptions(di), "TestCustomOptions"),
	)
}

func TestRealServerMWMockingViaMocker(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		WithMockedMiddleware(
			// Custom Mocker is required for real server, Option 1, directly provide a MWMocker
			MWCustomMocker(MWMockFunc(realServerMockFunc)),
		),
		apptest.WithModules(basicauth.Module, access.Module, errorhandling.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Invoke(registerTestController, registerTestSecurity),
		),
		test.GomegaSubTest(SubTestSuccessWithRealServer(di), "TestSuccessWithRealServer"),
		test.GomegaSubTest(SubTestFailedWithRealServer(di), "TestFailedWithRealServer"),
	)
}

func TestRealServerMWMockingViaConfigurer(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		WithMockedMiddleware(
			// Custom Mocker is required for real server, Option 2, via security.Configurer
			MWCustomConfigurer(security.ConfigurerFunc(realServerSecConfigurer)),
		),
		apptest.WithModules(basicauth.Module, access.Module, errorhandling.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Invoke(registerTestController, registerTestSecurity),
		),
		test.GomegaSubTest(SubTestSuccessWithRealServer(di), "TestSuccessWithRealServer"),
		test.GomegaSubTest(SubTestFailedWithRealServer(di), "TestFailedWithRealServer"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestWithoutMockedContext(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = webtest.NewRequest(ctx, http.MethodGet, TestSecuredURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp).To(Not(BeNil()), "response shouldn't be nil")
		g.Expect(resp.StatusCode).To(BeNumerically(">=", 300), "response should be > 300")
		g.Expect(resp.StatusCode).To(BeNumerically("<=", 399), "response should be <= 399")
	}
}

func SubTestWithMockedContext(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx = WithMockedSecurity(ctx, securityMock())
		var req *http.Request
		var resp *http.Response
		req = webtest.NewRequest(ctx, http.MethodGet, TestSecuredURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp).To(Not(BeNil()), "response shouldn't be nil")
		g.Expect(resp.StatusCode).To(BeNumerically("==", 200), "response should be 200")
	}
}

func SubTestCustomOptions(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		// GET, condition matched
		req = webtest.NewRequest(ctx, http.MethodGet, TestSecuredURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp).To(Not(BeNil()), "response shouldn't be nil")
		g.Expect(resp.StatusCode).To(BeNumerically("==", 200), "response should be 200")

		// POST, condition not matched
		req = webtest.NewRequest(ctx, http.MethodPost, TestSecuredURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp).To(Not(BeNil()), "response shouldn't be nil")
		g.Expect(resp.StatusCode).To(BeNumerically(">=", 300), "response should be > 300")
		g.Expect(resp.StatusCode).To(BeNumerically("<=", 399), "response should be <= 399")
	}
}

func SubTestSuccessWithRealServer(_ *testDI) test.GomegaSubTestFunc {
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

func SubTestFailedWithRealServer(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = webtest.NewRequest(ctx, http.MethodGet, TestSecuredURL, nil)
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp).To(Not(BeNil()), "response shouldn't be nil")
		// Note, in real server case, redirection is automatically followed
		g.Expect(resp.StatusCode).To(BeNumerically("==", 404), "response should be 404")
		g.Expect(resp.Request.URL.RequestURI()).To(HaveSuffix(TestEntryPointURL), "response's request path should be entry point URL")
	}
}

/*************************
	Helpers
 *************************/

func securityMock() SecurityMockOptions {
	return func(d *SecurityDetailsMock) {
		d.Username = "test-user"
		d.UserId = uuid.New().String()
		d.TenantId = uuid.New().String()
		d.TenantExternalId = "test-tenant"
		d.Permissions = utils.NewStringSet("TEST_PERMISSION")
	}
}

func newCustomMocker() MWMocker {
	return MWMockFunc(customMockFunc())
}

func customMockFunc() func(mc MWMockContext) security.Authentication {
	return func(mc MWMockContext) security.Authentication {
		return mockedAuth{}
	}
}

func realServerMockFunc(mc MWMockContext) security.Authentication {
	if mc.Request.Header.Get(TestHeader) == "" {
		return nil
	}
	return mockedAuth{}
}

func realServerSecConfigurer(ws security.WebSecurity) {
	ws = ws.Route(matcher.AnyRoute()).
		With(NewMockedMW().
			Mocker(MWMockFunc(realServerMockFunc)),
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
