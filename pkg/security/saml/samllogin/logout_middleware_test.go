package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	lanaisaml "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
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
	TestContextPath      = webtest.DefaultContextPath
	TestLogoutURL        = "/logout"
	TestLogoutSuccessURL = "/logout/success"
	TestLogoutErrorURL   = "/logout/error"
)

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
	warnings := logout.GetWarnings(ctx)
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
	logoutHandler logout.LogoutHandler
}

func (c *TestSecConfigurer) Configure(ws security.WebSecurity) {

	ws.Route(matcher.RouteWithPattern("/**")).
		With(logout.New().
			LogoutUrl(TestLogoutURL).
			SuccessUrl(TestLogoutSuccessURL).
			ErrorUrl(TestLogoutErrorURL).
			SuccessHandler(WarningsAwareSuccessHandler(TestLogoutSuccessURL)),
		).
		With(NewLogout().
			Issuer(testdata.TestIssuer).
			ErrorPath(TestLogoutErrorURL),
		)
}

type LogoutTestOut struct {
	fx.Out
	SecConfigurer security.Configurer
	IdpManager    idp.IdentityProviderManager
	AccountStore  security.FederatedAccountStore
}

func LogoutTestSecurityConfigProvider(registrar security.Registrar) LogoutTestOut {
	cfg := TestSecConfigurer{}
	registrar.Register(&cfg)
	return LogoutTestOut{
		SecConfigurer: &cfg,
		IdpManager:    testdata.NewTestIdpManager(),
		AccountStore:  testdata.NewTestFedAccountStore(),
	}
}

func ResetMocks(di *testDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		return ctx, nil
	}
}


/*************************
	Test
 *************************/

type testDI struct {
	fx.In
}

func TestSingleLogout(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithModules(SamlAuthModule, security.Module, logout.Module, csrf.Module, request_cache.Module, lanaisaml.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(LogoutTestSecurityConfigProvider),
		),
		test.SubTestSetup(ResetMocks(di)),
		//test.GomegaSubTest(SubTestLogoutError(di), "TestLogoutError"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLogoutError(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response

		req = webtest.NewRequest(ctx, http.MethodGet, TestLogoutURL, nil)
		req.Header.Set("Content-Type", "application/json")
		resp = webtest.MustExec(ctx, req).Response
		assertErrorResponse(t, g, resp)
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
	g.Expect(loc.RequestURI()).To(BeEquivalentTo(TestContextPath+TestLogoutSuccessURL), "response should redirect to success URL")
}

func assertEntryPointResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	loc := assertRedirectResponse(t, g, resp)
	g.Expect(loc.RequestURI()).To(BeEquivalentTo(TestContextPath+""), "response should redirect to entry point URL")
}

func assertErrorResponse(t *testing.T, g *gomega.WithT, resp *http.Response) {
	loc := assertRedirectResponse(t, g, resp)
	g.Expect(loc.RequestURI()).To(BeEquivalentTo(TestContextPath+TestLogoutErrorURL), "response should redirect to error URL")
}

func assertSuccessWithWarningResponse(t *testing.T, g *gomega.WithT, resp *http.Response, expectedWarning string) {
	loc := assertRedirectResponse(t, g, resp)
	g.Expect(loc.RequestURI()).To(HavePrefix(TestContextPath+TestLogoutSuccessURL), "response should redirect to success URL")
	g.Expect(loc.Query().Get("warning")).To(BeEquivalentTo(expectedWarning), "redirect URL should contains correct warning query")

}
