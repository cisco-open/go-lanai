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
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/utils/gomega"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/crewjam/saml"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// go:embed testdata/template/*.tmpl
//var TestHTMLTemplates embed.FS

/*************************
	Setup
 *************************/

const (
	ExpectedRedirectError     = `/test/error`
	ExpectedAuthorizeCallback = `http://localhost/test/callback`
	TestClientID              = "test-client"
	TestClientSecret          = "test-secret"
	TestOAuth2CallbackURL     = "http://localhost/oauth/callback"
)

// TestMain is the only place we should kick off embedded redis
func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		embedded.Redis(),
	)
}

type IntegrationTestDI struct {
	fx.In
	AppCtx  *bootstrap.ApplicationContext
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

		FedAccountStore: sectest.NewMockedFederatedAccountStore(testdata.MapValues(di.Mocking.FedAccounts)...),
		SamlClientStore: samltest.NewMockedClientStore(samltest.ClientsWithPropertiesPrefix(di.AppCtx.Config(), "mocking.clients")),
	}
}

/*************************
	Test
 *************************/

type intDI struct {
	fx.In
	FedAccountStore security.FederatedAccountStore
	Mocking         testdata.MockingProperties
	TokenReader     oauth2.TokenStoreReader
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
		//token tests
		test.GomegaSubTest(SubTestOAuth2AuthCode(di), "TestOAuth2AuthCode"),
		test.GomegaSubTest(SubTestOAuth2AuthCodeWithoutTenant(di), "TestOAuth2AuthCodeWithoutTenant"),

		//switch tenant

		// a user has access to two tenants, switch from one to the other
		//  the permission is not per tenant, so user permission doesn't change
		test.GomegaSubTest(SubTestOAuth2SwitchTenant(di), "TestOauth2SwitchTenant"),

		// a user that doesn't have access to any tenant adds a tenant, switch to this new tenant
		//  the permission is per tenant, so user permission is changed
		test.GomegaSubTest(SubTestOauth2SwitchToJustCreatedTenant(di), "TestOauth2SwitchToJustCreatedTenant"),

		// a user has access to two tenants, switch from one to the other
		//  the permission is per tenant, so user permission is changed
		test.GomegaSubTest(SubTestOauth2SwitchTenantWithPerTenantPermission(di), "TestOauth2SwitchTenantWithPerTenantPermission"),
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
			withDefaultAuthCode())
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
			withDefaultAuthCode())
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

func SubTestOAuth2AuthCode(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// mock authentication
		fedAccount := di.Mocking.FedAccounts["fed1"]
		ctx, e := contextWithSamlAuth(ctx, di.FedAccountStore, fedAccount)
		g.Expect(e).To(Succeed(), "SAML auth should be stored correctly")

		// authorize
		req := webtest.NewRequest(ctx, http.MethodGet, "/v2/authorize", nil, authorizeReqOptions())
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusFound), "response should have correct status code")
		assertAuthorizeResponse(t, g, resp.Response, false)

		// token
		code := extractAuthCode(resp.Response)
		req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token", authCodeReqBody(code), tokenReqOptions())
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")
		a := assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)

		auth, e := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)
		userDetail, ok := auth.Details().(security.UserDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(userDetail.UserId()).To(Equal(fedAccount.UserId))
		tenantDetail, ok := auth.Details().(security.TenantDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(tenantDetail.TenantId()).To(Equal(fedAccount.DefaultTenant))
		providerDetail, ok := auth.Details().(security.ProviderDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(providerDetail.ProviderId()).To(Not(BeEmpty()))
	}
}

func SubTestOAuth2AuthCodeWithoutTenant(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// mock authentication
		fedAccount := di.Mocking.FedAccounts["fed2"]
		ctx, e := contextWithSamlAuth(ctx, di.FedAccountStore, fedAccount)
		g.Expect(e).To(Succeed(), "SAML auth should be stored correctly")

		// authorize
		req := webtest.NewRequest(ctx, http.MethodGet, "/v2/authorize", nil, authorizeReqOptions())
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusFound), "response should have correct status code")
		assertAuthorizeResponse(t, g, resp.Response, false)

		// token
		code := extractAuthCode(resp.Response)
		req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token", authCodeReqBody(code), tokenReqOptions())
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")
		a := assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)

		auth, e := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)
		userDetail, ok := auth.Details().(security.UserDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(userDetail.UserId()).To(Equal(fedAccount.UserId))
		tenantDetail, ok := auth.Details().(security.TenantDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(tenantDetail.TenantId()).To(BeEmpty())
		providerDetail, ok := auth.Details().(security.ProviderDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(providerDetail.ProviderId()).To(BeEmpty())
	}
}

func SubTestOAuth2SwitchTenant(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// mock authentication
		fedAccount := di.Mocking.FedAccounts["fed1"]
		ctx, e := contextWithSamlAuth(ctx, di.FedAccountStore, fedAccount)
		g.Expect(e).To(Succeed(), "SAML auth should be stored correctly")

		// authorize
		req := webtest.NewRequest(ctx, http.MethodGet, "/v2/authorize", nil, authorizeReqOptions())
		resp := webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusFound), "response should have correct status code")
		assertAuthorizeResponse(t, g, resp.Response, false)

		// token
		code := extractAuthCode(resp.Response)
		req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token", authCodeReqBody(code), tokenReqOptions())
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK), "response should have correct status code")
		a := assertTokenResponse(t, g, resp.Response, fedAccount.Username, true)

		//switch tenant
		req = webtest.NewRequest(ctx, http.MethodPost, "/v2/token", switchTenantBody(a.Value(), "id-tenant-2"), tokenReqOptions(), withDefaultClientAuth())
		resp = webtest.MustExec(ctx, req)
		g.Expect(resp).ToNot(BeNil())
		g.Expect(resp.Response.StatusCode).To(Equal(http.StatusOK))
		a = assertTokenResponse(t, g, resp.Response, fedAccount.Username, false)
		auth, e := di.TokenReader.ReadAuthentication(ctx, a.Value(), oauth2.TokenHintAccessToken)

		tenantDetails, ok := auth.Details().(security.TenantDetails)
		g.Expect(ok).To(BeTrue())
		g.Expect(tenantDetails.TenantId()).To(Equal("id-tenant-2"))
	}
}

// TODO notes:
// The existing implementation of switching tenant is the following (See switch_tenant.go)
// 1. user has security.SpecialPermissionSwitchTenant - TODO: either we make this configurable, or in CDA every user needs to have this permission seeded.
// 2. reduce scope (if the requested client is different, or if request indicated different scope) - we don't need to worry about this part.
// 3. switch authentication
//   3.a create context details:
//       this will load new facts: account, tenant, verify tenant access, load provider - TODO: this is where change needs to be made, we need to load per tenant permission and compliment/override the global user permission
//   3.b create user auth based on context details - TODO: this is also where change needs to be made, when we switch tenant, we didn't use to switch user permission, so we re-use the old userAuth, but now that needs to change.
//       we need to do something similar to switch user (i.e. loadUserAuthentication)
//
// The general approach should be that whenever we are loading tenant, we need to give a chance to override the user's permission
// in case the permission is per tenanted.

func SubTestOauth2SwitchToJustCreatedTenant(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {

	}
}

func SubTestOauth2SwitchTenantWithPerTenantPermission(di *intDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {

	}
}

/*************************
	Helpers
 *************************/

func withClientAuth(clientId, secret string) webtest.RequestOptions {
	v := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientId, secret)))
	return webtest.Headers("Authorization", fmt.Sprintf("Basic %s", v))
}

func withDefaultClientAuth() webtest.RequestOptions {
	return withClientAuth(TestClientID, TestClientSecret)
}

func withDefaultAuthCode() webtest.RequestOptions {
	return webtest.Queries(
		oauth2.ParameterGrantType, oauth2.GrantTypeAuthCode,
		oauth2.ParameterClientId, TestClientID,
		oauth2.ParameterRedirectUri, TestOAuth2CallbackURL,
	)
}

func withDefaultSamlSSO() webtest.RequestOptions {
	return func(req *http.Request) {
		webtest.Queries(
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

func contextWithSamlAuth(ctx context.Context, fedAcctStore security.FederatedAccountStore, mock *sectest.MockedFederatedUserProperties) (context.Context, error) {
	assertion := mockAssertion(mock)
	acct, e := fedAcctStore.LoadAccountByExternalId(ctx,
		mock.ExtIdName,
		mock.ExtIdValue,
		mock.ExtIdpName,
		MockAutoCreateUserDetails{},
		assertion)
	if e != nil {
		return nil, e
	}

	return sectest.ContextWithSecurity(ctx, sectest.Authentication(
		&samltest.MockedSamlAssertionAuthentication{
			Account:       acct,
			DetailsMap:    map[string]interface{}{},
			SamlAssertion: assertion,
		}),
	), nil
}

func mockAssertion(mock *sectest.MockedFederatedUserProperties) *saml.Assertion {
	return samltest.MockAssertion(func(opt *samltest.AssertionOption) {
		opt.NameIDFormat = "urn:oasis:names:tc:SAML:1.1:nameid-format:email"
		opt.NameID = mock.ExtIdValue
		opt.RequestID = uuid.New().String()
		opt.Issuer = "http://some-entity-id"
		opt.Recipient = "http://some-sp/sso"
		opt.Audience = "http://some-sp"
	})
}

type MockAutoCreateUserDetails struct{}

func (m MockAutoCreateUserDetails) IsEnabled() bool {
	return true
}

func (m MockAutoCreateUserDetails) GetEmailWhiteList() []string {
	return []string{}
}

func (m MockAutoCreateUserDetails) GetAttributeMapping() map[string]string {
	return map[string]string{
		"firstName": "FirstName",
		"lastName":  "LastName",
		"email":     "Email",
	}
}

func (m MockAutoCreateUserDetails) GetElevatedUserRoleNames() []string {
	return []string{}
}

func (m MockAutoCreateUserDetails) GetRegularUserRoleNames() []string {
	return []string{}
}

func authorizeReqOptions() webtest.RequestOptions {
	return func(req *http.Request) {
		req.Host = fmt.Sprintf("http://%s", testdata.IdpDomainExtSAML)
		req.URL.Host = fmt.Sprintf("http://%s", testdata.IdpDomainExtSAML)
		values := url.Values{}
		values.Set(oauth2.ParameterGrantType, oauth2.GrantTypeAuthCode)
		values.Set(oauth2.ParameterResponseType, "code")
		values.Set(oauth2.ParameterClientId, "test-client")
		values.Set(oauth2.ParameterRedirectUri, "http://localhost/test/callback")
		req.URL.RawQuery = values.Encode()
	}
}

func extractAuthCode(resp *http.Response) string {
	loc := resp.Header.Get("Location")
	locUrl, _ := url.Parse(loc)
	return locUrl.Query().Get("code")
}

func authCodeReqBody(code string) io.Reader {
	values := url.Values{}
	values.Set(oauth2.ParameterGrantType, oauth2.GrantTypeAuthCode)
	values.Set(oauth2.ParameterClientId, "test-client")
	values.Set(oauth2.ParameterClientSecret, "test-secret")
	values.Set(oauth2.ParameterRedirectUri, "http://localhost/test/callback")
	values.Set(oauth2.ParameterAuthCode, code)
	return strings.NewReader(values.Encode())
}

func tokenReqOptions() webtest.RequestOptions {
	return func(req *http.Request) {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
	}
}

func switchTenantBody(accessToken string, tenantId string) io.Reader {
	values := url.Values{}
	values.Set(oauth2.ParameterGrantType, oauth2.GrantTypeSwitchTenant)
	values.Set(oauth2.ParameterAccessToken, accessToken)
	values.Set(oauth2.ParameterTenantId, tenantId)
	return strings.NewReader(values.Encode())
}

func assertTokenResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedUsername string, expectRefreshToken bool) oauth2.AccessToken {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `token response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$.access_token"), "token response should have access_token")
	g.Expect(body).To(HaveJsonPath("$.expires_in"), "token response should have expires_in")
	g.Expect(body).To(HaveJsonPath("$.scope"), "token response should have scope")
	g.Expect(body).To(HaveJsonPathWithValue("$.token_type", ContainElement("bearer")), "token response should have token_type")
	g.Expect(body).To(HaveJsonPathWithValue("$.username", expectedUsername), "token response should have correct username")

	if expectRefreshToken {
		g.Expect(body).To(HaveJsonPath("$.refresh_token"), "token response should have refresh_token")
	} else {
		g.Expect(body).NotTo(HaveJsonPath("$..refresh_token"), "token response should not have refresh_token")
	}

	accessToken := oauth2.NewDefaultAccessToken("")
	e = json.Unmarshal(body, accessToken)
	g.Expect(e).ToNot(HaveOccurred())
	return accessToken
}

func assertAuthorizeResponse(t *testing.T, g *gomega.WithT, resp *http.Response, expectErr bool) {
	g.Expect(resp.Header.Get("Set-Cookie")).To(Not(BeEmpty()), "authorize response should set cookie")
	switch {
	case expectErr:
		g.Expect(resp.Header.Get("Location")).To(Equal(ExpectedRedirectError), "authorize response should redirect to error page")
	default:
		assertCallbackRedirectResponse(t, g, resp)
	}
}

func assertCallbackRedirectResponse(_ *testing.T, g *gomega.WithT, resp *http.Response) {
	expected, _ := url.Parse(ExpectedAuthorizeCallback)
	loc := resp.Header.Get("Location")
	locUrl, e := url.Parse(loc)
	g.Expect(e).To(Succeed(), "authorize redirect location should be a valid URL")
	g.Expect(locUrl.Scheme).To(Equal(expected.Scheme), "authorize redirect should have correct scheme")
	g.Expect(locUrl.Host).To(Equal(expected.Host), "authorize redirect should have correct host")
	g.Expect(locUrl.Path).To(Equal(expected.Path), "authorize redirect should have correct path")
	q := locUrl.Query()
	g.Expect(q.Get("code")).To(Not(BeEmpty()), "authorize redirect queries should have code")
	return
}
