package basicauth_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"encoding/base64"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"testing"
)

/*************************
	Test Setup
 *************************/

const (
	TestUser1 = `test-user-1`
	TestUser1Password = `TheCakeIsALie`
	TestUser2 = `test-user-2`
)

type cfgDI struct {
	fx.In
	SecRegistrar security.Registrar
	WebRegistrar *web.Registrar
	AccountStore *sectest.MockAccountStore
}

func ConfigureTestWithBasicAuth(di cfgDI) {
	di.WebRegistrar.MustRegister(TestController{})
	di.SecRegistrar.Register(security.ConfigurerFunc(func(ws security.WebSecurity) {
		ws = ws.Route(matcher.RouteWithPattern("/secured/**")).
			With(passwd.New().
				MFA(false).AccountStore(di.AccountStore),
			).
			With(access.New().
				Request(matcher.AnyRequest()).Authenticated(),
			).
			With(errorhandling.New())
		basicauth.Configure(ws).
			EntryPoint(basicauth.NewBasicAuthEntryPoint())
	}))
}

func NewMockedAccountStore(accts sectest.MockedPropertiesAccounts) *sectest.MockAccountStore {
	return sectest.NewMockedAccountStore(accts.Values())
}

/*************************
	Tests
 *************************/

type AuthDI struct {
	fx.In
}

func TestAuthenticator(t *testing.T) {
	var di AuthDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithModules(security.Module, access.Module, errorhandling.Module, passwd.Module, basicauth.Module),
		apptest.WithFxOptions(
			fx.Provide(
                sectest.MockedPropertiesBinder[sectest.MockedPropertiesAccounts]("accounts"),
                NewMockedAccountStore,
			),
			fx.Invoke(ConfigureTestWithBasicAuth),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestAuthSuccess(&di), "AuthSuccess"),
		test.GomegaSubTest(SubTestWrongPassword(&di), "WrongPassword"),
		test.GomegaSubTest(SubTestMissingHeader(&di), "MissingHeader"),
		test.GomegaSubTest(SubTestWrongTokenType(&di), "WrongHeaderType"),
		test.GomegaSubTest(SubTestMalformedEncoding(&di), "MalformedEncoding"),
		test.GomegaSubTest(SubTestSameUser(&di), "SameUser"),
		test.GomegaSubTest(SubTestDifferentUser(&di), "DifferentUser"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestAuthSuccess(_ *AuthDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, TestUser1, TestUser1Password)
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusOK)
	}
}

func SubTestWrongPassword(_ *AuthDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, TestUser1, "hah?")
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusUnauthorized)
	}
}

func SubTestMissingHeader(_ *AuthDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, "", "")
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusUnauthorized)
	}
}

func SubTestWrongTokenType(_ *AuthDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, "", "")
		resp = webtest.MustExec(ctx, req, webtest.Headers("Authorization", "Bearer some-token-value")).Response
		AssertResponse(g, resp, http.StatusUnauthorized)
	}
}

func SubTestMalformedEncoding(_ *AuthDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, "", "")
		resp = webtest.MustExec(ctx, req, webtest.Headers("Authorization", "Basic some-token-value")).Response
		AssertResponse(g, resp, http.StatusUnauthorized)

		basic := base64.StdEncoding.EncodeToString([]byte("user-without-password-portion"))
		req = NewRequest(ctx, "", "")
		resp = webtest.MustExec(ctx, req, webtest.Headers("Authorization", "Basic "+basic)).Response
		AssertResponse(g, resp, http.StatusUnauthorized)
	}
}

func SubTestSameUser(_ *AuthDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		auth := MockedUserAuth{
			Authentication: sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption) {
				opt.Principal = TestUser1
				opt.State = security.StateAuthenticated
			}),
		}
		ctx = sectest.ContextWithSecurity(ctx, sectest.Authentication(auth))
		req = NewRequest(ctx, TestUser1, "")
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusOK)
	}
}

func SubTestDifferentUser(_ *AuthDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		auth := MockedUserAuth{
			Authentication: sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption) {
				opt.Principal = TestUser2
				opt.State = security.StateAuthenticated
			}),
		}
		ctx = sectest.ContextWithSecurity(ctx, sectest.Authentication(auth))
		req = NewRequest(ctx, TestUser1, "")
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusUnauthorized)
	}
}

/*************************
	Helper
 *************************/

func NewRequest(ctx context.Context, username, password string) *http.Request {
	headers := []string{"Accept", "application/json"}
	if len(username) != 0 {
		basic := fmt.Sprintf("%s:%s", username, password)
		basic = base64.StdEncoding.EncodeToString([]byte(basic))
		headers = append(headers, "Authorization", "Basic "+basic)
	}
	return webtest.NewRequest(ctx, http.MethodPost, "/secured/post", nil, webtest.Headers(headers...))
}

func AssertResponse(g *gomega.WithT, resp *http.Response, expectedSC int) {
	g.Expect(resp).ToNot(BeNil(), "response should not be nil")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "response status code should be correct")
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), "reading response body should not fail")
	g.Expect(body).ToNot(BeEmpty(), "response body should not be empty")
	if expectedSC == http.StatusOK {
		return
	}

	if expectedSC == http.StatusUnauthorized || expectedSC == http.StatusForbidden {
		g.Expect(resp.Header.Get("WWW-Authenticate")).To(HavePrefix("Basic "), "response should have correct '%s' header", "WWW-Authenticate")
	}
}

type TestController struct{}

func (c TestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Post("/secured/post").EndpointFunc(c.Post).Build(),
	}
}

func (c TestController) Post(_ context.Context, _ *http.Request) (interface{}, error) {
	return map[string]interface{}{
		"success": "yay",
	}, nil
}

type MockedUserAuth struct {
	security.Authentication
}

func (a MockedUserAuth) Username() string {
	return a.Authentication.Principal().(string)
}

func (a MockedUserAuth) IsMFAPending() bool {
	return false
}

func (a MockedUserAuth) OTPIdentifier() string {
	return ""
}


