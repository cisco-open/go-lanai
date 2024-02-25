package formlogin_test

import (
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/formlogin"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
    "cto-github.cisco.com/NFV-BU/go-lanai/test"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
    "embed"
    "errors"
    "fmt"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    gotemplate "html/template"
    "net/http"
    "regexp"
    "strings"
    "testing"
)

/*************************
   Test Setup
*************************/

//go:embed testdata/*.tmpl
var TemplateFS embed.FS

func FormLoginPageTestConfigurer() func(di cfgDI) {
	return func(di cfgDI) {
		di.WebRegistrar.MustRegister(NewDefaultLoginFormController())
		di.WebRegistrar.MustRegister(TemplateFS)
		_ = di.WebRegistrar.AddEngineOptions(func(eng *web.Engine) {
			eng.SetFuncMap(gotemplate.FuncMap{
				"printKV": PrintKV,
			})
		})
	}
}

/*************************
   Tests
*************************/

type PageDI struct {
	fx.In
	SessionStore     session.Store
	AccountStore     *sectest.MockAccountStore
	MFAEventRecorder *MFAEventRecorder `optional:"true"`
}

func TestFormLoginController(t *testing.T) {
	var di PageDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(security.Module, access.Module, errorhandling.Module,
			passwd.Module, formlogin.Module, csrf.Module),
		apptest.WithFxOptions(
			fx.Provide(
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesAccounts]("accounts"),
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesTenants]("tenants"),
				NewMockedAccountStoreWithMFA,
			),
			fx.Invoke(FormLoginTestConfigurer(true)),
			fx.Invoke(FormLoginPageTestConfigurer()),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestLoginPageWithError(&di), "LoginPageWithError"),
		test.GomegaSubTest(SubTestLoginPageWithRememberMe(&di), "LoginPageWithRememberMe"),
		test.GomegaSubTest(SubTestOTPPageWithError(&di), "OTPPageWithError"),
		test.GomegaSubTest(SubTestOTPPageWithoutAuth(&di), "OTPPageWithoutAuth"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLoginPageWithError(di *PageDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
        ctx, s:= ContextWithSession(ctx, di.SessionStore)
        s.AddFlash(errors.New("error message"), redirect.FlashKeyPreviousError)
        s.AddFlash(TestUser1, ParamUsername)

		req = NewFormRequestWithSession(ctx, s, http.MethodGet, UrlLoginPage + "?error=true")
		resp = webtest.MustExec(ctx, req).Response
		AssertHtmlResponse(g, resp, http.StatusOK,
            `error`, "error message",
			`csrf`, `[a-zA-Z0-9\-]+`,
            `usernameParam`, ParamUsername,
            `passwordParam`, ParamPassword,
            `loginProcessUrl`, UrlLoginProcess,
            ParamUsername, TestUser1,
		)
	}
}

func SubTestLoginPageWithRememberMe(di *PageDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        var req *http.Request
        var resp *http.Response
        ctx, s:= ContextWithSession(ctx, di.SessionStore)
        s.AddFlash(errors.New("error message"), redirect.FlashKeyPreviousError)
        e := s.Save()
        g.Expect(e).To(Succeed(), "save session should not fail")

        req = NewFormRequestWithSession(ctx, s, http.MethodGet, UrlLoginPage + "?error=true")
        resp = webtest.MustExec(ctx, req, func(req *http.Request) {
            cookie := session.NewCookie(formlogin.CookieKeyRememberedUsername, TestUser2, &session.Options{}, req)
            req.AddCookie(cookie)
        }).Response
        AssertHtmlResponse(g, resp, http.StatusOK,
            `error`, "error message",
            `csrf`, `[a-zA-Z0-9\-]+`,
            `usernameParam`, ParamUsername,
            `passwordParam`, ParamPassword,
            `loginProcessUrl`, UrlLoginProcess,
            `rememberedUsername`, TestUser2,
        )
    }
}

func SubTestOTPPageWithError(di *PageDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        var req *http.Request
        var resp *http.Response
        ctx, s:= ContextWithSession(ctx, di.SessionStore)
        ctx = sectest.ContextWithSecurity(ctx, sectest.Authentication(&MockedUserAuth{
            Authentication: sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption) {
                opt.Principal = TestUser1
                opt.State = security.StateAuthenticated
                opt.Permissions = map[string]interface{} {
                    passwd.SpecialPermissionMFAPending: true,
                    passwd.SpecialPermissionOtpId: "test-otp-id",
                }
            }),
            OTP:            "test-otp-value",
        }))
        s.AddFlash(errors.New("error message"), redirect.FlashKeyPreviousError)

        req = NewFormRequestWithSession(ctx, s, http.MethodGet, UrlOTPPage + "?error=true")
        resp = webtest.MustExec(ctx, req).Response
        AssertHtmlResponse(g, resp, http.StatusOK,
            `error`, "error message",
            `csrf`, `[a-zA-Z0-9\-]+`,
            `otpParam`, ParamOTP,
            `mfaVerifyUrl`, UrlOTPProcess,
            `mfaRefreshUrl`, UrlOTPRefresh,
        )
    }
}

func SubTestOTPPageWithoutAuth(di *PageDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        var req *http.Request
        var resp *http.Response
        ctx, s:= ContextWithSession(ctx, di.SessionStore)
        req = NewFormRequestWithSession(ctx, s, http.MethodGet, UrlOTPPage + "?error=true")
        resp = webtest.MustExec(ctx, req).Response
        AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlLoginPage))
    }
}

/*************************
	Helpers
 *************************/

func PrintKV(model map[string]any) string {
	lines := make([]string, 0, len(model))
	for k, v := range model {
		var line string
        switch v := v.(type) {
        case *csrf.Token:
            line = fmt.Sprintf(`%s=%v`, k, v.Value)
        default:
            line = fmt.Sprintf(`%s=%v`, k, v)
        }

		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func AssertHtmlResponse(g *gomega.WithT, resp *http.Response, expectedSC int, kvs ...string) {
	body := AssertResponse(g, resp, expectedSC)
	g.Expect(body).ToNot(BeEmpty(), "response body should be empty")
	for i := 1; i < len(kvs); i += 2 {
		pattern := fmt.Sprintf(`%s=%v`, regexp.QuoteMeta(kvs[i-1]), kvs[i])
		g.Expect(string(body)).To(MatchRegexp(pattern), "response body should be match pattern [%s]", pattern)
	}
}
