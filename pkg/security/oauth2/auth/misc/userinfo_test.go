package misc_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/misc"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http/httptest"
	"testing"
	"time"
)

/*************************
	Setup Test
 *************************/

var MockedSecurityDetailsTmpl = sectest.SecurityDetailsMock{
	Username:                 "test-user",
	UserId:                   "test-id",
	TenantExternalId:         "root-tenant",
	TenantId:                 "root-tenant-id",
	ProviderName:             "test-provider",
	ProviderId:               "test-provider-id",
	ProviderDisplayName:      "Test Provider",
	ProviderDescription:      "Test Provider",
	ProviderEmail:            "test@email.com",
	ProviderNotificationType: "email",
	Exp:                      time.Now().Add(time.Minute),
	Iss:                      time.Now(),
	Tenants:                  utils.NewStringSet("root-tenant-id"),
	UserFirstName:            "John",
	UserLastName:             "Smith",
	KVs: map[string]interface{}{
		"Email": "test-user@test.com",
	},
	ClientID: ClientIDMinor,
	Scopes:   utils.NewStringSet("read", "write"),
}

/*************************
	Test
 *************************/

type UserInfoDI struct {
	fx.In
	AuthDI
	Endpoint     *misc.UserInfoEndpoint
	AccountStore security.AccountStore
	JwtDecoder   jwt.JwtDecoder
}

func TestUserInfoEndpoint(t *testing.T) {
	var di UserInfoDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithFxOptions(
			fx.Provide(
				sectest.BindMockingProperties,
				NewTestIssuer, NewTestClientStore,
				NewTestAccountStore,
				NewJwkStore, NewJwtEncoder, NewJwtDecoder,
			),
			fx.Provide(misc.NewUserInfoEndpoint),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestJwtUserInfo(&di), "JwtUserInfo"),
		test.GomegaSubTest(SubTestPlainUserInfo(&di), "PlainUserInfo"),
		test.GomegaSubTest(SubTestUserInfoWithOIDCScope(&di), "UserInfoWithOIDCScope"),
		test.GomegaSubTest(SubTestUserInfoWithClaimsRequest(&di), "TestUserInfoWithClaimsRequest"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestJwtUserInfo(di *UserInfoDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const username = TestUser1
		var req misc.UserInfoRequest
		var resp misc.UserInfoJwtResponse
		var e error
		// no security
		resp, e = di.Endpoint.JwtUserInfo(ctx, req)
		g.Expect(e).To(HaveOccurred(), "JwtUserInfo should fail without authentication")

		// with security
		acct, e := di.AccountStore.LoadAccountByUsername(ctx, username)
		g.Expect(e).To(Succeed(), "load account [%s] should not fail", username)
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(SecurityMockWithAccount(acct)))
		resp, e = di.Endpoint.JwtUserInfo(ctx, req)
		g.Expect(e).To(Succeed(), "JwtUserInfo should not fail")
		AssertUserInfoJwt(g, string(resp), di.JwtDecoder, acct, ExpectUserInfoProfile, ExpectUserInfoEmail, ExpectUserInfoPhone, ExpectUserInfoAddress)

		// verify custom encoder
		respWriter := httptest.NewRecorder()
		e = misc.JwtResponseEncoder()(ctx, respWriter, resp)
		g.Expect(e).To(Succeed(), "encoding JWT response should not fail")
		g.Expect(respWriter.Header().Get("Content-Type")).To(HavePrefix("application/jwt"), "JWT response should have correct content-type")
		g.Expect(respWriter.Body.Bytes()).To(Equal([]byte(resp)), "JWT response should have correct body")
	}
}

func SubTestPlainUserInfo(di *UserInfoDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const username = TestUser1
		var req misc.UserInfoRequest
		var resp *misc.UserInfoPlainResponse
		var e error
		// no security
		resp, e = di.Endpoint.PlainUserInfo(ctx, req)
		g.Expect(e).To(HaveOccurred(), "PlainUserInfo should fail without authentication")

		// withSecurity
		acct, e := di.AccountStore.LoadAccountByUsername(ctx, username)
		g.Expect(e).To(Succeed(), "load account [%s] should not fail", username)
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(SecurityMockWithAccount(acct)))
		resp, e = di.Endpoint.PlainUserInfo(ctx, req)
		g.Expect(e).To(Succeed(), "PlainUserInfo should not fail")
		AssertUserInfoClaims(g, resp, acct, ExpectUserInfoProfile, ExpectUserInfoEmail, ExpectUserInfoPhone, ExpectUserInfoAddress)
	}
}

func SubTestUserInfoWithOIDCScope(di *UserInfoDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const username = TestUser1
		var req misc.UserInfoRequest
		var resp *misc.UserInfoPlainResponse
		var e error
		acct, e := di.AccountStore.LoadAccountByUsername(ctx, username)
		g.Expect(e).To(Succeed(), "load account [%s] should not fail", username)
		// without extra profiles
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(
			SecurityMockWithAccount(acct),
			func(d *sectest.SecurityDetailsMock) {
				d.Scopes.Add(oauth2.ScopeOidc)
			},
		))
		resp, e = di.Endpoint.PlainUserInfo(ctx, req)
		g.Expect(e).To(Succeed(), "PlainUserInfo should not fail")
		AssertUserInfoClaims(g, resp, acct, ExpectUserInfoNoProfile, ExpectUserInfoNoEmail)

		// with extra profiles
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(
			SecurityMockWithAccount(acct),
			func(d *sectest.SecurityDetailsMock) {
				d.Scopes.Add(oauth2.ScopeOidc, oauth2.ScopeOidcProfile)
			},
		))
		resp, e = di.Endpoint.PlainUserInfo(ctx, req)
		g.Expect(e).To(Succeed(), "PlainUserInfo should not fail")
		AssertUserInfoClaims(g, resp, acct, ExpectUserInfoProfile, ExpectUserInfoNoEmail)
	}
}

func SubTestUserInfoWithClaimsRequest(di *UserInfoDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const username = TestUser1
		var req misc.UserInfoRequest
		var resp *misc.UserInfoPlainResponse
		var e error
		var claimsReq string
		acct, e := di.AccountStore.LoadAccountByUsername(ctx, username)
		g.Expect(e).To(Succeed(), "load account [%s] should not fail", username)

		// essential
		claimsReq = `{"userinfo":{"email":{"essential":true},"email_verified":{"essential":true}}}`
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(
			SecurityMockWithAccount(acct),
			func(d *sectest.SecurityDetailsMock) {
				d.Scopes.Add(oauth2.ScopeOidc)
				d.OAuth2Parameters = map[string]string{
					oauth2.ParameterClaims: claimsReq,
				}
			},
		))
		resp, e = di.Endpoint.PlainUserInfo(ctx, req)
		g.Expect(e).To(Succeed(), "PlainUserInfo should not fail")
		AssertUserInfoClaims(g, resp, acct, ExpectUserInfoNoProfile, ExpectUserInfoEmail)

		// non-essential, no scope
		claimsReq = `{"userinfo":{"email":null,"email_verified":null}}`
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(
			SecurityMockWithAccount(acct),
			func(d *sectest.SecurityDetailsMock) {
				d.Scopes.Add(oauth2.ScopeOidc)
				d.OAuth2Parameters = map[string]string{
					oauth2.ParameterClaims: claimsReq,
				}
			},
		))
		resp, e = di.Endpoint.PlainUserInfo(ctx, req)
		g.Expect(e).To(Succeed(), "PlainUserInfo should not fail")
		AssertUserInfoClaims(g, resp, acct, ExpectUserInfoNoProfile, ExpectUserInfoNoEmail)

		// non-essential, with scope scope
		claimsReq = `{"userinfo":{"email":null,"email_verified":null}}`
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(
			SecurityMockWithAccount(acct),
			func(d *sectest.SecurityDetailsMock) {
				d.Scopes.Add(oauth2.ScopeOidc, oauth2.ScopeOidcEmail)
				d.OAuth2Parameters = map[string]string{
					oauth2.ParameterClaims: claimsReq,
				}
			},
		))
		resp, e = di.Endpoint.PlainUserInfo(ctx, req)
		g.Expect(e).To(Succeed(), "PlainUserInfo should not fail")
		AssertUserInfoClaims(g, resp, acct, ExpectUserInfoNoProfile, ExpectUserInfoEmail)
	}
}

/*************************
	Helpers
 *************************/

func SecurityMockWithAccount(acct security.Account) sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		*d = MockedSecurityDetailsTmpl
		d.Username = acct.Username()
		d.UserId = fmt.Sprintf(`%v`, acct.ID())
		d.TenantId = acct.(security.AccountTenancy).TenantId()
		d.Permissions = utils.NewStringSetFrom(d.Permissions)
		d.Permissions.Add(acct.Permissions()...)
		d.Tenants = utils.NewStringSetFrom(d.Tenants)
		d.Tenants.Add(acct.(security.AccountTenancy).DesignatedTenantIds()...)
		d.Scopes = utils.NewStringSetFrom(d.Scopes)
		kvs := map[string]interface{}{}
		for k, v := range d.KVs {
			kvs[k] = v
		}
		d.KVs = kvs
	}
}

type ExpectedUserInfoSpecs func(acct security.Account) []ExpectedClaimsOption

func AssertUserInfoJwt(g *gomega.WithT, jwtValue string, decoder jwt.JwtDecoder, expectAcct security.Account, expectedSpecs ...ExpectedUserInfoSpecs) {
	var claims misc.UserInfoClaims
	e := decoder.DecodeWithClaims(context.Background(), jwtValue, &claims)
	g.Expect(e).To(Succeed(), "decoding JWT should not fail")
	AssertUserInfoClaims(g, &claims, expectAcct, expectedSpecs...)
}

func AssertUserInfoClaims(g *gomega.WithT, claims oauth2.Claims, expectAcct security.Account, expectedSpecs ...ExpectedUserInfoSpecs) {
	expectOpts := []ExpectedClaimsOption{
		ExpectClaim(oauth2.ClaimIssuer, "http://"+IssuerDomain+IssuerPath),
		ExpectClaim(oauth2.ClaimSubject, expectAcct.Username()),
		ExpectClaim(oauth2.ClaimAudience, utils.NewStringSet(ClientIDMinor)),
	}
	for _, specs := range expectedSpecs {
		expectOpts = append(expectOpts, specs(expectAcct)...)
	}
	AssertClaims(g, claims, NewExpectedClaims(expectOpts...))
}

func ExpectUserInfoProfile(acct security.Account) []ExpectedClaimsOption {
	return []ExpectedClaimsOption {
		ExpectClaim(oauth2.ClaimFullName, Not(BeZero())),
		ExpectClaim(oauth2.ClaimFirstName, Not(BeZero())),
		ExpectClaim(oauth2.ClaimLastName, Not(BeZero())),
		ExpectClaim(oauth2.ClaimPermissions, utils.NewStringSet(acct.Permissions()...)),
		ExpectClaim(oauth2.ClaimPreferredUsername, acct.Username()),
		ExpectClaim(oauth2.ClaimDefaultTenantId, acct.(security.AccountTenancy).DefaultDesignatedTenantId()),
	}
}

func ExpectUserInfoNoProfile(_ security.Account) []ExpectedClaimsOption {
	return []ExpectedClaimsOption {
		ExpectClaim(oauth2.ClaimFullName, nil),
		ExpectClaim(oauth2.ClaimFirstName, nil),
		ExpectClaim(oauth2.ClaimLastName, nil),
		ExpectClaim(oauth2.ClaimPermissions, nil),
		ExpectClaim(oauth2.ClaimPreferredUsername, nil),
		ExpectClaim(oauth2.ClaimDefaultTenantId, nil),
	}
}

func ExpectUserInfoEmail(_ security.Account) []ExpectedClaimsOption {
	return []ExpectedClaimsOption {
		ExpectClaim(oauth2.ClaimEmail, Not(BeZero())),
		ExpectClaim(oauth2.ClaimEmailVerified, &utils.TRUE),
	}
}

func ExpectUserInfoNoEmail(_ security.Account) []ExpectedClaimsOption {
	return []ExpectedClaimsOption {
		ExpectClaim(oauth2.ClaimEmail, nil),
		ExpectClaim(oauth2.ClaimEmailVerified, nil),
	}
}

func ExpectUserInfoPhone(_ security.Account) []ExpectedClaimsOption {
	return []ExpectedClaimsOption {
		// currently, security.Account doesn't support phone
	}
}

func ExpectUserInfoAddress(_ security.Account) []ExpectedClaimsOption {
	return []ExpectedClaimsOption {
		// currently, security.Account doesn't support address
	}
}
