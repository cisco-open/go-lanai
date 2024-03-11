package seclient_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/integrate/httpclient"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/seclient"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sdtest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

const (
	SDAuthServiceName      = `mocked-auth-server`
	TestClientID           = `test-client`
	TestClientSecret       = `test-secret`
	TestAccountID          = `test-account-id`
	TestAccount            = `test-account`
	TestPassword           = `test-password`
	TestTenantID           = `test-tenant-id`
	TestTenantName         = `test-tenant`
	TestScope              = `test-scope`
	TestAltClientID        = `test-client-alt`
	TestAltClientSecret    = `test-secret-alt`
	TestCurrentAccessToken = `test-token`
)

/*************************
	Tests
 *************************/

type TestDI struct {
	fx.In
	sdtest.DI
	AuthClient       seclient.AuthenticationClient
	MockedController *MockedController
}

func TestWithMockedServer(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),
		apptest.WithModules(httpclient.Module, seclient.Module),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(NewMockedController),
			web.FxControllerProviders(ProvideWebController),
		),
		test.SubTestSetup(UpdateMockedSD(&di)),
		test.GomegaSubTest(SubTestClientCredentials(&di), "ClientCredentials"),
		test.GomegaSubTest(SubTestPasswordLogin(&di), "PasswordLogin"),
		test.GomegaSubTest(SubTestSwitchUser(&di), "SwitchUser"),
		test.GomegaSubTest(SubTestSwitchTenant(&di), "SwitchTenant"),
	)
}

/*************************
	Sub Tests
 *************************/

// UpdateMockedSD update SD record to use the random server port
func UpdateMockedSD(di *TestDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		port := webtest.CurrentPort(ctx)
		if port <= 0 {
			return ctx, nil
		}
		di.Client.UpdateMockedService(SDAuthServiceName, sdtest.AnyInstance(), func(inst *discovery.Instance) {
			inst.Port = port
		})
		return ctx, nil
	}
}

func SubTestClientCredentials(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		test.RunTest(ctx, t,
			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// without options
				rs, e := di.AuthClient.ClientCredentials(ctx)
				g.Expect(e).To(Succeed(), "authentication should not fail")
				AssertResult(g, rs, oauth2.GrantTypeClientCredentials, TestClientID, TestClientSecret)
			}, "MinimumOptions"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// with more options
				rs, e := di.AuthClient.ClientCredentials(ctx, seclient.WithClientAuth(TestAltClientID, TestAltClientSecret),
					seclient.WithScopes(TestScope), seclient.WithTenantId(TestTenantID))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeClientCredentials, TestAltClientID, TestAltClientSecret)
				g.Expect(req.Scopes).To(HaveKey(TestScope), "request's [%s] should be correct", "Scopes")
				AssertTokenRequestParams(g, req, oauth2.ParameterTenantId, TestTenantID)
			}, "MoreOptions"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// failure
				_, e := di.AuthClient.ClientCredentials(ctx, seclient.WithClientAuth(InvalidClientID, "whatever"))
				g.Expect(e).To(HaveOccurred(), "authentication should fail with invalid credentials")
			}, "Failure"),
		)
	}
}

func SubTestPasswordLogin(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		test.RunTest(ctx, t,
			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// minimum options
				rs, e := di.AuthClient.PasswordLogin(ctx, seclient.WithCredentials(TestAccount, TestPassword))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypePassword, TestClientID, TestClientSecret)
				AssertTokenRequestParams(g, req,
					oauth2.ParameterUsername, TestAccount, oauth2.ParameterPassword, TestPassword, oauth2.ParameterNonce, Not(BeEmpty()))
			}, "MinimumOptions"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// with more options
				rs, e := di.AuthClient.PasswordLogin(ctx, seclient.WithCredentials(TestAccount, TestPassword),
					seclient.WithCurrentSecurity(ctx),
					seclient.WithClientAuth(TestAltClientID, TestAltClientSecret),
					seclient.WithScopes(TestScope), seclient.WithTenantId(TestTenantID),
				)
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypePassword, TestAltClientID, TestAltClientSecret)
				g.Expect(req.Scopes).To(HaveKey(TestScope), "request's [%s] should be correct", "Scopes")
				AssertTokenRequestParams(g, req,
					oauth2.ParameterUsername, TestAccount, oauth2.ParameterPassword, TestPassword, oauth2.ParameterNonce, Not(BeEmpty()),
					oauth2.ParameterTenantId, TestTenantID, oauth2.ParameterAccessToken, nil)
			}, "MoreOptions"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// failure
				_, e := di.AuthClient.PasswordLogin(ctx, seclient.WithCredentials(TestAccount, TestPassword), seclient.WithClientAuth(InvalidClientID, "whatever"))
				g.Expect(e).To(HaveOccurred(), "authentication should fail with invalid credentials")
			}, "Failure"),
		)
	}
}

func SubTestSwitchUser(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		test.RunTest(ctx, t,
			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// minimum options - with username
				rs, e := di.AuthClient.SwitchUser(ctx, seclient.WithAccessToken(TestCurrentAccessToken), seclient.WithUsername(TestAccount))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchUser, TestClientID, TestClientSecret)
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterSwitchUsername, TestAccount, oauth2.ParameterNonce, Not(BeEmpty()))
			}, "WithUsername"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// minimum options - with user ID
				rs, e := di.AuthClient.SwitchUser(ctx, seclient.WithAccessToken(TestCurrentAccessToken), seclient.WithUserId(TestAccountID))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchUser, TestClientID, TestClientSecret)
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterSwitchUserId, TestAccountID, oauth2.ParameterNonce, Not(BeEmpty()))
			}, "WithUserID"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// minimum options - with
				rs, e := di.AuthClient.SwitchUser(ctx, seclient.WithAccessToken(TestCurrentAccessToken),
					seclient.WithUser(TestAccountID, TestAccount))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchUser, TestClientID, TestClientSecret)
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterSwitchUsername, TestAccount, oauth2.ParameterNonce, Not(BeEmpty()))
			}, "WithUsernameAndID"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// with more options
				rs, e := di.AuthClient.SwitchUser(ctx, seclient.WithAccessToken(TestCurrentAccessToken),
					seclient.WithUsername(TestAccount),
					seclient.WithClientAuth(TestAltClientID, TestAltClientSecret),
					seclient.WithScopes(TestScope), seclient.WithTenantId(TestTenantID),
				)
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchUser, TestAltClientID, TestAltClientSecret)
				g.Expect(req.Scopes).To(HaveKey(TestScope), "request's [%s] should be correct", "Scopes")
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterSwitchUsername, TestAccount, oauth2.ParameterNonce, Not(BeEmpty()),
					oauth2.ParameterTenantId, TestTenantID)
			}, "WithMoreOptions"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// current security
				ctx = ContextWithSecurity(ctx)
				rs, e := di.AuthClient.SwitchUser(ctx, seclient.WithCurrentSecurity(ctx), seclient.WithUsername(TestAccount))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchUser, TestClientID, TestClientSecret)
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterSwitchUsername, TestAccount, oauth2.ParameterNonce, Not(BeEmpty()))
			}, "WithCurrentSecurity"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// failure
				_, e := di.AuthClient.SwitchUser(ctx, seclient.WithClientAuth(InvalidClientID, "whatever"))
				g.Expect(e).To(HaveOccurred(), "authentication should fail with invalid credentials")
			}, "Failure"),
		)
	}
}

func SubTestSwitchTenant(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		test.RunTest(ctx, t,
			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// minimum options - with tenant id
				rs, e := di.AuthClient.SwitchTenant(ctx, seclient.WithAccessToken(TestCurrentAccessToken),
					seclient.WithTenantId(TestTenantID))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchTenant, TestClientID, TestClientSecret)
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterTenantId, TestTenantID, oauth2.ParameterNonce, Not(BeEmpty()))
			}, "WithTenantID"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// minimum options - with user ID
				rs, e := di.AuthClient.SwitchTenant(ctx, seclient.WithAccessToken(TestCurrentAccessToken),
					seclient.WithTenantExternalId(TestTenantName))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchTenant, TestClientID, TestClientSecret)
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterTenantExternalId, TestTenantName, oauth2.ParameterNonce, Not(BeEmpty()))
			}, "WithTenantName"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// minimum options - with
				rs, e := di.AuthClient.SwitchTenant(ctx, seclient.WithAccessToken(TestCurrentAccessToken),
					seclient.WithTenant(TestTenantID, TestTenantName))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchTenant, TestClientID, TestClientSecret)
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterTenantId, TestTenantID, oauth2.ParameterNonce, Not(BeEmpty()))
			}, "WithTenantIDAndName"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// with more options
				rs, e := di.AuthClient.SwitchTenant(ctx, seclient.WithAccessToken(TestCurrentAccessToken),
					seclient.WithTenantId(TestTenantID),
					seclient.WithClientAuth(TestAltClientID, TestAltClientSecret),
					seclient.WithScopes(TestScope),
				)
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchTenant, TestAltClientID, TestAltClientSecret)
				g.Expect(req.Scopes).To(HaveKey(TestScope), "request's [%s] should be correct", "Scopes")
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterTenantId, TestTenantID, oauth2.ParameterNonce, Not(BeEmpty()), )
			}, "WithMoreOptions"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// current security
				ctx = ContextWithSecurity(ctx)
				rs, e := di.AuthClient.SwitchTenant(ctx, seclient.WithCurrentSecurity(ctx), seclient.WithTenantId(TestTenantID))
				g.Expect(e).To(Succeed(), "authentication should not fail")
				req := AssertResult(g, rs, oauth2.GrantTypeSwitchTenant, TestClientID, TestClientSecret)
				AssertTokenRequestParams(g, req,
					oauth2.ParameterAccessToken, TestCurrentAccessToken, oauth2.ParameterTenantId, TestTenantID, oauth2.ParameterNonce, Not(BeEmpty()))
			}, "WithCurrentSecurity"),

			test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *WithT) {
				// failure
				_, e := di.AuthClient.SwitchTenant(ctx, seclient.WithClientAuth(InvalidClientID, "whatever"))
				g.Expect(e).To(HaveOccurred(), "authentication should fail with invalid credentials")
			}, "Failure"),
		)
	}
}

/*************************
	Helpers
 *************************/

func AssertResult(g *gomega.WithT, rs *seclient.Result, expectedGrant, expectedClientId, expectedSecret string) *auth.TokenRequest {
	g.Expect(rs).ToNot(BeNil(), "result should be available")
	g.Expect(rs.Token).ToNot(BeNil(), "result token should be available")
	g.Expect(rs.Token.Value()).ToNot(BeNil(), "result token value should be available")

	req, e := DecodeMockedAccessTokenValue(rs.Token.Value())
	g.Expect(e).To(Succeed(), "decoding result token value should not fail")
	g.Expect(req).ToNot(BeNil(), "decoded request should not be nil")
	g.Expect(req.GrantType).To(Equal(expectedGrant), "request's [%s] should be correct", "GrantType")
	if len(expectedClientId) != 0 {
		g.Expect(req.ClientId).To(Equal(expectedClientId), "request's [%s] should be correct", "ClientId")
		AssertTokenRequestExtensions(g, req, ExtClientSecret, expectedSecret)
	}
	return req
}

func AssertTokenRequestParams(g *gomega.WithT, req *auth.TokenRequest, expectedKVs ...interface{}) {
	for i := 1; i < len(expectedKVs); i += 2 {
		if expectedKVs[i] != nil {
			g.Expect(req.Parameters).To(HaveKeyWithValue(expectedKVs[i-1], expectedKVs[i]),
				"request's parameters should have correct [%s]", expectedKVs[i-1])
		} else {
			g.Expect(req.Parameters).ToNot(HaveKey(expectedKVs[i-1]), "request's parameters should not have [%s]", expectedKVs[i-1])
		}
	}
}

func AssertTokenRequestExtensions(g *gomega.WithT, req *auth.TokenRequest, expectedKVs ...interface{}) {
	for i := 1; i < len(expectedKVs); i += 2 {
		if expectedKVs[i] != nil {
			g.Expect(req.Extensions).To(HaveKeyWithValue(expectedKVs[i-1], expectedKVs[i]),
				"request's extensions should have correct [%s]", expectedKVs[i-1])
		} else {
			g.Expect(req.Extensions).ToNot(HaveKey(expectedKVs[i-1]), "request's extensions should not have [%s]", expectedKVs[i-1])
		}
	}
}

func ContextWithSecurity(ctx context.Context) context.Context {
	return sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.AccessToken = TestCurrentAccessToken
	}))
}