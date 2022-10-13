package config

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/basicauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/authserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/resserver"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/config/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/formlogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/extsamlidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp/passwdidp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/logout"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/authorize"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/clientauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/token"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	samlidp "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/idp"
	samlsp "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/sp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/embedded"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samltest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"encoding/base64"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"net/url"
	"testing"
)

// go:embed testdata/template/*.tmpl
//var TestHTMLTemplates embed.FS

/*************************
	Setup
 *************************/

const (
	TestClientID          = "test-client"
	TestClientSecret      = "test-secret"
	TestOAuth2CallbackURL = "http://localhost/oauth/callback"
)

// TestMain is the only place we should kick off embedded redis
func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		embedded.Redis(),
	)
}

type IntegrationTestDI struct {
	fx.In
	AppCtx *bootstrap.ApplicationContext
	SecReg  security.Registrar
	WebReg  *web.Registrar
	Mocking testdata.MockingProperties
}

type IntegrationTestOut struct {
	fx.Out
	DiscoveryCustomizers *discovery.Customizers
	IdpManager           idp.IdentityProviderManager
	AccountStore         security.AccountStore
	PasswordEncoder      passwd.PasswordEncoder
	FedAccountStore      security.FederatedAccountStore
	SamlClientStore      samlctx.SamlClientStore
}

func IntegrationTestMocksProvider(di IntegrationTestDI) IntegrationTestOut {
	return IntegrationTestOut{
		DiscoveryCustomizers: &discovery.Customizers{Customizers: utils.NewSet()},
		IdpManager:           testdata.NewMockedIDPManager(),
		AccountStore:         sectest.NewMockedAccountStore(testdata.MapValues(di.Mocking.Accounts)...),
		PasswordEncoder:      passwd.NewNoopPasswordEncoder(),
		FedAccountStore:      sectest.NewMockedFederatedAccountStore(),
		SamlClientStore:      samltest.NewMockedClientStore(samltest.ClientsWithPropertiesPrefix(di.AppCtx.Config(), "mocking.clients")),
	}
}

/*************************
	Test
 *************************/

type intDI struct {
	fx.In
}

func TestWithMockedServer(t *testing.T) {
	di := &intDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(
			authserver.Module, resserver.Module,
			passwdidp.Module, extsamlidp.Module, authorize.Module, samlidp.Module,
			passwd.Module, formlogin.Module, logout.Module,
			samlctx.Module, samlsp.Module,
			basicauth.Module, clientauth.Module,
			token.Module, access.Module, errorhandling.Module,
			request_cache.Module, csrf.Module, session.Module,
			redis.Module,
		),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(
				IntegrationTestMocksProvider,
				testdata.BindMockingProperties,
				testdata.NewAuthServerConfigurer,
				testdata.NewResServerConfigurer,
			),
		),
		test.GomegaSubTest(SubTestOAuth2AuthorizeWithPasswdIDP(di), "TestOAuth2AuthorizeWithPasswdIDP"),
		test.GomegaSubTest(SubTestOAuth2AuthorizeWithSamlSSO(di), "TestOAuth2AuthorizeWithSamlSSO"),
		test.GomegaSubTest(SubTestSamlSSOAuthorizeWithPasswdIDP(di), "TestSamlSSOAuthorizeWithPasswdIDP"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestOAuth2AuthorizeWithPasswdIDP(_ *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		uri := fmt.Sprintf("http://%s/test/v2/authorize", testdata.IdpDomainPasswd)
		req = webtest.NewRequest(ctx, http.MethodGet, uri, nil,
			withDefaultClientAuth(), withDefaultAuthCode())
		resp = webtest.MustExec(ctx, req).Response
		fmt.Printf("%v\n", resp)
		assertRedirectResponse(t, g, resp, "/test/login")
	}
}

func SubTestOAuth2AuthorizeWithSamlSSO(_ *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		uri := fmt.Sprintf("http://%s/test/v2/authorize", testdata.IdpDomainExtSAML)
		req = webtest.NewRequest(ctx, http.MethodGet, uri, nil,
			withDefaultClientAuth(), withDefaultAuthCode())
		resp = webtest.MustExec(ctx, req).Response
		fmt.Printf("%v\n", resp)
		assertRedirectResponse(t, g, resp, testdata.ExtSamlIdpSSOUrl)
	}
}

func SubTestSamlSSOAuthorizeWithPasswdIDP(_ *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		uri := fmt.Sprintf("http://%s/test/v2/authorize", testdata.IdpDomainPasswd)
		req = webtest.NewRequest(ctx, http.MethodPost, uri, nil, withDefaultSamlSSO())
		resp = webtest.MustExec(ctx, req).Response
		fmt.Printf("%v\n", resp)
		assertRedirectResponse(t, g, resp, "/test/login")
	}
}

/*************************
	Helpers
 *************************/

func withClientAuth(clientId, secret string) webtest.RequestOptions {
	v := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientId, secret)))
	return webtest.WithHeaders("Authorization", v)
}

func withDefaultClientAuth() webtest.RequestOptions {
	return withClientAuth(TestClientID, TestClientSecret)
}

func withDefaultAuthCode() webtest.RequestOptions {
	return webtest.WithQueries(
		oauth2.ParameterGrantType, oauth2.GrantTypeAuthCode,
		oauth2.ParameterClientId, TestClientID,
		oauth2.ParameterRedirectUri, TestOAuth2CallbackURL,
	)
}

func withDefaultSamlSSO() webtest.RequestOptions {
	return func(req *http.Request) {
		webtest.WithQueries(
			oauth2.ParameterGrantType, oauth2.GrantTypeSamlSSO,
			oauth2.ParameterClientId, TestClientID,
			oauth2.ParameterRedirectUri, TestOAuth2CallbackURL,
		)(req)
	}
}

func assertRedirectResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedUrl string) {
	g.Expect(resp.StatusCode).To(BeNumerically("~", 300, 399), "response should be redirect")
	g.Expect(resp.Header).To(HaveKey("Location"), "response should have redirect location")

	expected, e := url.Parse(expectedUrl)
	g.Expect(e).To(Succeed(), "expected URL should be valid")
	loc, e := url.Parse(resp.Header.Get("Location"))
	g.Expect(e).To(Succeed(), "response's redirect location should be valid")
	if expected.IsAbs() {
		g.Expect(loc.String()).To(HavePrefix(expectedUrl), "response's redirect location should have correct host, port and path")
	} else {
		g.Expect(loc.Path).To(Equal(expectedUrl), "response's redirect location should have correct path")
	}
}
