package tokenauth_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

/*************************
	Setup Test
 *************************/

const (
	TestUser1    = `test-user-1`
	TestUser2    = `test-user-2`
	TestTenantID = `id-root`
	TestScope    = `test`
)

func NewTestTokenStoreReader(accts sectest.MockedPropertiesAccounts, tenants sectest.MockedPropertiesTenants) oauth2.TokenStoreReader {
	return sectest.NewMockedTokenStoreReader(accts.MapValues(), tenants.MapValues())
}

type cfgDI struct {
	fx.In
	SecRegistrar     security.Registrar
	WebRegistrar     *web.Registrar
	TokenStoreReader oauth2.TokenStoreReader
}

func ConfigureMinimumCustomization(di cfgDI) {
	configurer := tokenauth.NewTokenAuthConfigurer(func(opt *tokenauth.TokenAuthOption) {
		opt.TokenStoreReader = di.TokenStoreReader
	})
	di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(tokenauth.FeatureId, configurer)
	di.WebRegistrar.MustRegister(TestController{})
	di.SecRegistrar.Register(security.ConfigurerFunc(func(ws security.WebSecurity) {
		ws.Route(matcher.RouteWithPattern("/secured/**")).
			With(tokenauth.New()).
			With(access.New().
				Request(matcher.AnyRequest()).Authenticated(),
			).
			With(errorhandling.New())
	}))
}

func ConfigureCustomTokenAuth(di cfgDI) {
	configurer := tokenauth.NewTokenAuthConfigurer(func(opt *tokenauth.TokenAuthOption) {
		opt.TokenStoreReader = di.TokenStoreReader
	})
	di.SecRegistrar.(security.FeatureRegistrar).RegisterFeature(tokenauth.FeatureId, configurer)
	di.WebRegistrar.MustRegister(TestController{})
	di.SecRegistrar.Register(security.ConfigurerFunc(func(ws security.WebSecurity) {
		ws = ws.Route(matcher.RouteWithPattern("/secured/**")).
			With(access.New().
				Request(matcher.AnyRequest()).Authenticated(),
			).
			With(errorhandling.New())
		tokenauth.Configure(ws).
			EnablePostBody().
			ErrorHandler(tokenauth.NewOAuth2ErrorHanlder())
	}))
}

/*************************
	Test
 *************************/

type mwDI struct {
	fx.In
	TokenStoreReader oauth2.TokenStoreReader
}

func TestMiddlewareNoCustomization(t *testing.T) {
	var di mwDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithModules(security.Module, access.Module, errorhandling.Module, tokenauth.Module),
		apptest.WithFxOptions(
			fx.Provide(
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesAccounts]("accounts"),
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesTenants]("tenants"),
				NewTestTokenStoreReader,
			),
			fx.Invoke(ConfigureMinimumCustomization),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestValidToken(&di), "ValidToken"),
		test.GomegaSubTest(SubTestValidTokenNoScopes(&di), "ValidTokenNoScopes"),
		test.GomegaSubTest(SubTestInvalidToken(&di), "InvalidToken"),
		test.GomegaSubTest(SubTestRevokedToken(&di), "RevokedToken"),
		test.GomegaSubTest(SubTestMissingToken(&di), "MissingToken"),
		test.GomegaSubTest(SubTestWithExistingAuth(&di), "WithExistingAuth"),
		test.GomegaSubTest(SubTestValidTokenViaForm(&di, false), "ValidTokenViaForm"),
	)
}

func TestMiddlewareCustomized(t *testing.T) {
	var di mwDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithModules(security.Module, access.Module, errorhandling.Module, tokenauth.Module),
		apptest.WithFxOptions(
			fx.Provide(
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesAccounts]("accounts"),
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesTenants]("tenants"),
				NewTestTokenStoreReader,
			),
			fx.Invoke(ConfigureCustomTokenAuth),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestValidToken(&di), "ValidToken"),
		test.GomegaSubTest(SubTestValidTokenNoScopes(&di), "ValidTokenNoScopes"),
		test.GomegaSubTest(SubTestInvalidToken(&di), "InvalidToken"),
		test.GomegaSubTest(SubTestRevokedToken(&di), "RevokedToken"),
		test.GomegaSubTest(SubTestMissingToken(&di), "MissingToken"),
		test.GomegaSubTest(SubTestWithExistingAuth(&di), "WithExistingAuth"),
		test.GomegaSubTest(SubTestValidTokenViaForm(&di, true), "ValidTokenViaForm"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestValidToken(_ *mwDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, MockedTokenValue(TestUser1, TestTenantID, time.Minute, TestScope))
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusOK, false)
	}
}

func SubTestValidTokenNoScopes(_ *mwDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, MockedTokenValue(TestUser1, TestTenantID, time.Minute))
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusForbidden, false)
	}
}

func SubTestInvalidToken(_ *mwDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, "invalid-token")
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusUnauthorized, true)
	}
}

func SubTestRevokedToken(di *mwDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		token := MockedTokenValue(TestUser1, TestTenantID, time.Minute)
		di.TokenStoreReader.(sectest.MockedTokenRevoker).Revoke(token)
		req = NewRequest(ctx, token)
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusUnauthorized, true)
	}
}

func SubTestMissingToken(_ *mwDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, "")
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusUnauthorized, false)
	}
}

func SubTestWithExistingAuth(_ *mwDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
			d.Scopes = utils.NewStringSet(TestScope)
		}))
		req = NewRequest(ctx, "")
		resp = webtest.MustExec(ctx, req).Response
		AssertResponse(g, resp, http.StatusUnauthorized, false)
	}
}

func SubTestValidTokenViaForm(_ *mwDI, expectSuccess bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = NewRequest(ctx, "")
		resp = webtest.MustExec(ctx, req,
			func(req *http.Request) {
				kvs := fmt.Sprintf("%s=%s", "access_token",  MockedTokenValue(TestUser1, TestTenantID, time.Minute, TestScope))
				req.Body = io.NopCloser(strings.NewReader(kvs))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			},
		).Response
		if expectSuccess {
			AssertResponse(g, resp, http.StatusOK, false)
		} else {
			AssertResponse(g, resp, http.StatusUnauthorized, false)
		}
	}
}

/*************************
	Helper
 *************************/

func NewRequest(ctx context.Context, tokenValue string) *http.Request {
	headers := []string{"Accept", "application/json"}
	if len(tokenValue) != 0 {
		headers = append(headers, "Authorization", "Bearer "+tokenValue)
	}
	return webtest.NewRequest(ctx, http.MethodPost, "/secured/post", nil, webtest.Headers(headers...))
}

func MockedTokenValue(username, tenantId string, exp time.Duration, scopes ...string) string {
	now := time.Now()
	expTime := now.Add(exp)
	t := sectest.MockedToken{
		MockedTokenInfo: sectest.MockedTokenInfo{
			UName:  username,
			TID:    tenantId,
			Exp:    now.Unix(),
			Iss:    now.Unix(),
			Scopes: append([]string{"read", "write"}, scopes...),
		},
		ExpTime: expTime,
		IssTime: now,
	}
	text, e := t.MarshalText()
	if e != nil {
		return ""
	}
	return string(text)
}

func AssertResponse(g *gomega.WithT, resp *http.Response, expectedSC int, expectExtraHeaders bool) {
	g.Expect(resp).ToNot(BeNil(), "response should not be nil")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "response status code should be correct")
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), "reading response body should not fail")
	g.Expect(body).ToNot(BeEmpty(), "response body should not be empty")
	if expectedSC == http.StatusOK {
		return
	}

	if expectExtraHeaders {
		g.Expect(resp.Header.Get("Cache-Control")).To(Equal("no-store"), "response should have correct '%s' header", "Cache-Control")
		g.Expect(resp.Header.Get("Pragma")).To(Equal("no-cache"), "response should have correct '%s' header", "Pragma")
		if expectedSC == http.StatusUnauthorized || expectedSC == http.StatusForbidden {
			g.Expect(resp.Header.Get("WWW-Authenticate")).To(HavePrefix("Bearer "), "response should have correct '%s' header", "WWW-Authenticate")
		}
	} else {
		g.Expect(resp.Header.Get("Cache-Control")).To(BeEmpty(), "response should have correct '%s' header", "Cache-Control")
		g.Expect(resp.Header.Get("Pragma")).To(BeEmpty(), "response should have correct '%s' header", "Pragma")
		g.Expect(resp.Header.Get("WWW-Authenticate")).To(BeEmpty(), "response should have correct '%s' header", "WWW-Authenticate")
	}
}

type TestController struct{}

func (c TestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Post("/secured/post").Condition(tokenauth.RequireScopes(TestScope)).EndpointFunc(c.Post).Build(),
	}
}

func (c TestController) Post(_ context.Context, _ *http.Request) (interface{}, error) {
	return map[string]interface{}{
		"success": "yay",
	}, nil
}
