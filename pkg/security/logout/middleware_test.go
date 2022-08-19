package logout

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout/mocks"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

/*************************
	Setup
 *************************/

const (
	TestContextPath = "/test"
	TestLogoutURL = "/logout"
	TestLogoutSuccessURL = "/logout/success"
	TestLogoutErrorURL = "/logout/error"
	TestLogoutEntryPointURL = "/logout/cancelled"

	MethodNameShouldLogout = "ShouldLogout"
	MethodNameHandleLogout = "HandleLogout"
)

type MockedLogoutHandler struct {
	mocks.ConditionalLogoutHandler
	mocks.LogoutHandler
}

type WarningsAwareSuccessHandler string

func (h WarningsAwareSuccessHandler) HandleAuthenticationSuccess(ctx context.Context, r *http.Request, rw http.ResponseWriter, _, _ security.Authentication) {
	redirectUrl := string(h)
	if contextPath, ok := ctx.Value(web.ContextKeyContextPath).(string); ok {
		redirectUrl = contextPath + redirectUrl
	}
	redirectUrl = h.appendWarnings(ctx, redirectUrl)
	http.Redirect(rw, r, redirectUrl, http.StatusFound)
	_, _ = rw.Write([]byte{})
}

func (h WarningsAwareSuccessHandler) appendWarnings(ctx context.Context, redirect string) string {
	warnings := GetWarnings(ctx)
	if len(warnings) == 0 {
		return redirect
	}

	redirectUrl, e := url.Parse(redirect)
	if e != nil {
		return redirect
	}

	q := redirectUrl.Query()
	for _, w := range warnings {
		q.Add("warning", w.Error())
	}
	redirectUrl.RawQuery = q.Encode()
	return redirectUrl.String()
}

type TestSecConfigurer struct {
	logoutHandler LogoutHandler
}

func (c *TestSecConfigurer) Configure(ws security.WebSecurity) {
	anotherHandler := &mocks.LogoutHandler{}
	anotherHandler.On(MethodNameHandleLogout, NonZero, NonZero, NonZero, NonZero).Return(nil)

	ws.Route(matcher.RouteWithPattern("/**")).
		With(New().
			LogoutUrl(TestLogoutURL).
			EntryPoint(redirect.NewRedirectWithRelativePath(TestLogoutEntryPointURL, false)).
			SuccessUrl(TestLogoutSuccessURL).
			ErrorUrl(TestLogoutErrorURL).
			SuccessHandler(WarningsAwareSuccessHandler(TestLogoutSuccessURL)).
			LogoutHandlers(c.logoutHandler).
			AddLogoutHandler(anotherHandler),
		)
}

func SecurityConfigProvider(registrar security.Registrar) (security.Configurer, *MockedLogoutHandler) {
	h := MockedLogoutHandler{}
	cfg := TestSecConfigurer{
		logoutHandler: &h,
	}
	registrar.Register(&cfg)
	return &cfg, &h
}

func ResetMocks(di *testDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		di.MockedHandler.LogoutHandler.Mock = mock.Mock{}
		di.MockedHandler.ConditionalLogoutHandler.Mock = mock.Mock{}
		return ctx, nil
	}
}

func MockLogoutHandler(h *MockedLogoutHandler, cancel bool, err error) {
	h.LogoutHandler.
		On(MethodNameHandleLogout, NonZero, NonZero, NonZero, NonZero).
		Return(err)

	if cancel {
		h.ConditionalLogoutHandler.
			On(MethodNameShouldLogout, NonZero, NonZero, NonZero, NonZero).
			Return(errors.New("cancelled"))
	} else {
		h.ConditionalLogoutHandler.
			On(MethodNameShouldLogout, NonZero, NonZero, NonZero, NonZero).
			Return(nil)
	}
}

/*************************
	Test
 *************************/

type testDI struct {
	fx.In
	MockedHandler *MockedLogoutHandler
}

func TestMockedTestServer(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(webtest.UseContextPath(TestContextPath)),
		apptest.WithModules(Module, security.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(SecurityConfigProvider),
		),
		test.SubTestSetup(ResetMocks(di)),
		test.GomegaSubTest(SubTestLogoutSuccess(di), "TestLogoutSuccess"),
		test.GomegaSubTest(SubTestLogoutError(di), "TestLogoutError"),
		test.GomegaSubTest(SubTestLogoutCancelled(di), "TestLogoutCancelled"),
		test.GomegaSubTest(SubTestLogoutSuccessWithWarnings(di), "TestLogoutSuccessWithWarnings"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLogoutSuccess(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		MockLogoutHandler(di.MockedHandler, false, nil)
		// GET
		req = webtest.NewRequest(ctx, http.MethodGet, TestLogoutURL, nil)
		req.Header.Set("Content-Type", "application/json")
		resp = webtest.MustExec(ctx, req).Response
		assertSuccessResponse(t, g, resp)

		// POST
		req = webtest.NewRequest(ctx, http.MethodPost, TestLogoutURL, nil)
		req.Header.Set("Content-Type", "application/json")
		resp = webtest.MustExec(ctx, req).Response
		assertSuccessResponse(t, g, resp)

		// PUT
		req = webtest.NewRequest(ctx, http.MethodPut, TestLogoutURL, nil)
		req.Header.Set("Content-Type", "application/json")
		resp = webtest.MustExec(ctx, req).Response
		assertSuccessResponse(t, g, resp)

		// DELETE
		req = webtest.NewRequest(ctx, http.MethodDelete, TestLogoutURL, nil)
		req.Header.Set("Content-Type", "application/json")
		resp = webtest.MustExec(ctx, req).Response
		assertSuccessResponse(t, g, resp)
	}
}

func SubTestLogoutError(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		MockLogoutHandler(di.MockedHandler, false, errors.New("logout failed"))

		req = webtest.NewRequest(ctx, http.MethodGet, TestLogoutURL, nil)
		req.Header.Set("Content-Type", "application/json")
		resp = webtest.MustExec(ctx, req).Response
		assertErrorResponse(t, g, resp)
	}
}

func SubTestLogoutCancelled(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		MockLogoutHandler(di.MockedHandler, true, nil)

		req = webtest.NewRequest(ctx, http.MethodGet, TestLogoutURL, nil)
		req.Header.Set("Content-Type", "application/json")
		resp = webtest.MustExec(ctx, req).Response
		assertEntryPointResponse(t, g, resp)

		di.MockedHandler.ConditionalLogoutHandler.AssertNumberOfCalls(t, MethodNameShouldLogout, 1)
		di.MockedHandler.LogoutHandler.AssertNumberOfCalls(t, MethodNameHandleLogout, 0)
	}
}

func SubTestLogoutSuccessWithWarnings(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const warning = "watning message"
		var req *http.Request
		var resp *http.Response
		MockLogoutHandler(di.MockedHandler, false, security.NewAuthenticationWarningError(warning))

		req = webtest.NewRequest(ctx, http.MethodGet, TestLogoutURL, nil)
		req.Header.Set("Content-Type", "application/json")
		resp = webtest.MustExec(ctx, req).Response
		assertSuccessWithWarningResponse(t, g, resp, warning)
	}
}

/*************************
	Helpers
 *************************/

var (
	NonZero = mock.MatchedBy(func(i interface{}) bool {
		return !reflect.ValueOf(i).IsZero()
	})
)

func assertRedirectResponse(_ *testing.T, g *gomega.WithT, resp *http.Response) *url.URL {
	g.Expect(resp).To(Not(BeNil()), "response shouldn't be nil")
	g.Expect(resp.StatusCode).To(BeNumerically(">=", 300), "status code should be >= 300")
	g.Expect(resp.StatusCode).To(BeNumerically("<=", 399), "status code should be <= 399")
	loc, e := resp.Location()
	g.Expect(e).To(Succeed(), "Location header should be a valid URL")
	return loc
}

func assertSuccessResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	loc := assertRedirectResponse(t, g, resp)
	g.Expect(loc.RequestURI()).To(BeEquivalentTo(TestContextPath + TestLogoutSuccessURL), "response should redirect to success URL")
}

func assertEntryPointResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	loc := assertRedirectResponse(t, g, resp)
	g.Expect(loc.RequestURI()).To(BeEquivalentTo(TestContextPath + TestLogoutEntryPointURL), "response should redirect to entry point URL")
}

func assertErrorResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	loc := assertRedirectResponse(t, g, resp)
	g.Expect(loc.RequestURI()).To(BeEquivalentTo(TestContextPath + TestLogoutErrorURL), "response should redirect to error URL")
}

func assertSuccessWithWarningResponse(t *testing.T, g *gomega.WithT, resp *http.Response, expectedWarning string) {
	loc := assertRedirectResponse(t, g, resp)
	g.Expect(loc.RequestURI()).To(HavePrefix(TestContextPath + TestLogoutSuccessURL), "response should redirect to success URL")
	g.Expect(loc.Query().Get("warning")).To(BeEquivalentTo(expectedWarning), "redirect URL should contains correct warning query")

}