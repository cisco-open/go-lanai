package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
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
	ClientID:         ClientIDMinor,
	Scopes:           utils.NewStringSet("read", "write", "openid"),
	OAuth2GrantType:  oauth2.GrantTypeAuthCode,
	OAuth2Parameters: map[string]string{},
	OAuth2Extensions: map[string]interface{}{},
}

func ProvideOpenIDTokenEnhancer(issuer security.Issuer, encoder jwt.JwtEncoder) *OpenIDTokenEnhancer {
	return NewOpenIDTokenEnhancer(func(opt *EnhancerOption) {
		opt.Issuer = issuer
		opt.JwtEncoder = encoder
	})
}

/*************************
	Test
 *************************/

type IDTokenDI struct {
	fx.In
	TokenEnhancer *OpenIDTokenEnhancer
	AccountStore  security.AccountStore
	JwtEncoder    jwt.JwtEncoder
	JwtDecoder    jwt.JwtDecoder
}

func TestOpenIDTokenEnhancer(t *testing.T) {
	var di IDTokenDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithFxOptions(
			fx.Provide(
				BindMockingProperties, NewTestIssuer, NewTestAccountStore,
				NewJwkStore, NewJwtEncoder, NewJwtDecoder,
				ProvideOpenIDTokenEnhancer,
			),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestIDTokenWithCodeRespType(&di), "IDTokenWithCodeRespType"),
		test.GomegaSubTest(SubTestIDTokenWithTokenRespType(&di), "IDTokenWithTokenRespType"),
		test.GomegaSubTest(SubTestIDTokenWithIDTokenRespType(&di), "IDTokenWithIDTokenRespType"),
		test.GomegaSubTest(SubTestIDTokenWithClaimsRequest(&di), "IDTokenWithClaimsRequest"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestIDTokenWithCodeRespType(di *IDTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const username = TestUser1
		var token oauth2.AccessToken
		var e error

		acct, e := di.AccountStore.LoadAccountByUsername(ctx, username)
		g.Expect(e).To(Succeed(), "load account [%s] should not fail", username)
		auth := OAuth2AuthenticationWithAccount(acct,
			func(d *sectest.SecurityDetailsMock) {
				d.AccessToken = MockedJWTValue(di.JwtEncoder)
				d.KVs[security.DetailsKeyAuthMethod] = security.AuthMethodPassword
				d.OAuth2Parameters[oauth2.ParameterNonce] = TestNonce
				d.OAuth2Parameters[oauth2.ParameterMaxAge] = "180"
				d.OAuth2ResponseTypes = utils.NewStringSet("code")
			},
		)
		token, e = di.TokenEnhancer.Enhance(ctx, oauth2.FromAccessToken(auth.AccessToken()), auth)
		g.Expect(e).To(Succeed(), "Enhance() should not fail")
		g.Expect(token).ToNot(BeNil(), "enhanced token should not be nil")
		g.Expect(token.Details()).To(HaveKeyWithValue("id_token", BeAssignableToTypeOf("")),
			"enhanced token should contains 'id_token'")
		AssertIDToken(g, token.Details()["id_token"].(string), di.JwtDecoder, acct, auth)
	}
}

func SubTestIDTokenWithTokenRespType(di *IDTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const username = TestUser1
		var token oauth2.AccessToken
		var e error
		var auth oauth2.Authentication
		acct, e := di.AccountStore.LoadAccountByUsername(ctx, username)
		g.Expect(e).To(Succeed(), "load account [%s] should not fail", username)
		// only "token"
		auth = OAuth2AuthenticationWithAccount(acct,
			func(d *sectest.SecurityDetailsMock) {
				d.AccessToken = MockedJWTValue(di.JwtEncoder)
				d.KVs[security.DetailsKeyAuthMethod] = security.AuthMethodPassword
				d.OAuth2Parameters[oauth2.ParameterNonce] = TestNonce
				d.OAuth2Parameters[oauth2.ParameterMaxAge] = "180"
				d.OAuth2ResponseTypes = utils.NewStringSet("token")
			},
		)
		token, e = di.TokenEnhancer.Enhance(ctx, oauth2.FromAccessToken(auth.AccessToken()), auth)
		g.Expect(e).To(Succeed(), "Enhance() should not fail")
		g.Expect(token).ToNot(BeNil(), "enhanced token should not be nil")
		g.Expect(token.Details()).ToNot(HaveKey("id_token"), "enhanced token should not contains 'id_token' when response_type='token'")

		// only "token id_token"
		auth = OAuth2AuthenticationWithAccount(acct,
			func(d *sectest.SecurityDetailsMock) {
				d.AccessToken = MockedJWTValue(di.JwtEncoder)
				d.KVs[security.DetailsKeyAuthMethod] = security.AuthMethodPassword
				d.OAuth2Parameters[oauth2.ParameterNonce] = TestNonce
				d.OAuth2Parameters[oauth2.ParameterMaxAge] = "180"
				d.OAuth2ResponseTypes = utils.NewStringSet("token", "id_token")
			},
		)
		token, e = di.TokenEnhancer.Enhance(ctx, oauth2.FromAccessToken(auth.AccessToken()), auth)
		g.Expect(e).To(Succeed(), "Enhance() should not fail")
		g.Expect(token).ToNot(BeNil(), "enhanced token should not be nil")
		g.Expect(token.Details()).To(HaveKeyWithValue("id_token", BeAssignableToTypeOf("")),
			"enhanced token should contains 'id_token' when response_type='token id_token'")
		AssertIDToken(g, token.Details()["id_token"].(string), di.JwtDecoder, acct, auth)
	}
}

func SubTestIDTokenWithIDTokenRespType(di *IDTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const username = TestUser1
		var token oauth2.AccessToken
		var e error

		acct, e := di.AccountStore.LoadAccountByUsername(ctx, username)
		g.Expect(e).To(Succeed(), "load account [%s] should not fail", username)
		auth := OAuth2AuthenticationWithAccount(acct,
			func(d *sectest.SecurityDetailsMock) {
				d.AccessToken = MockedJWTValue(di.JwtEncoder)
				d.KVs[security.DetailsKeyAuthMethod] = security.AuthMethodPassword
				d.OAuth2Parameters[oauth2.ParameterNonce] = TestNonce
				d.OAuth2Parameters[oauth2.ParameterMaxAge] = "180"
				d.OAuth2ResponseTypes = utils.NewStringSet("id_token")
			},
		)
		token, e = di.TokenEnhancer.Enhance(ctx, oauth2.FromAccessToken(auth.AccessToken()), auth)
		g.Expect(e).To(Succeed(), "Enhance() should not fail")
		g.Expect(token).ToNot(BeNil(), "enhanced token should not be nil")
		g.Expect(token.Details()).To(HaveKeyWithValue("id_token", BeAssignableToTypeOf("")),
			"enhanced token should contains 'id_token'")
		AssertIDToken(g, token.Details()["id_token"].(string), di.JwtDecoder, acct, auth)
	}
}

func SubTestIDTokenWithClaimsRequest(di *IDTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const username = TestUser1
		var token oauth2.AccessToken
		var e error
		var auth oauth2.Authentication
		acct, e := di.AccountStore.LoadAccountByUsername(ctx, username)
		g.Expect(e).To(Succeed(), "load account [%s] should not fail", username)
		// only "token"
		auth = OAuth2AuthenticationWithAccount(acct,
			func(d *sectest.SecurityDetailsMock) {
				d.AccessToken = MockedJWTValue(di.JwtEncoder)
				d.KVs[security.DetailsKeyAuthMethod] = security.AuthMethodPassword
				d.OAuth2Parameters[oauth2.ParameterNonce] = TestNonce
				d.OAuth2Parameters[oauth2.ParameterMaxAge] = "180"
				d.OAuth2Parameters[oauth2.ParameterClaims] = `{"id_token":{"email":{"essential":true}}}`
				d.OAuth2ResponseTypes = utils.NewStringSet("code")
			},
		)
		token, e = di.TokenEnhancer.Enhance(ctx, oauth2.FromAccessToken(auth.AccessToken()), auth)
		g.Expect(e).To(Succeed(), "Enhance() should not fail")
		g.Expect(token).ToNot(BeNil(), "enhanced token should not be nil")
		g.Expect(token.Details()).To(HaveKeyWithValue("id_token", BeAssignableToTypeOf("")),
			"enhanced token should contains 'id_token' when response_type='token id_token'")
		AssertIDToken(g, token.Details()["id_token"].(string), di.JwtDecoder, acct, auth,
			ExpectClaim(oauth2.ClaimEmail, Not(BeZero())),
		)
	}
}

/*************************
	Helpers
 *************************/

func MockedJWTValue(encoder jwt.JwtEncoder) string {
	claims := oauth2.MapClaims{
		oauth2.ClaimSubject: "doesn't matter",
	}
	v, _ := encoder.Encode(context.Background(), claims)
	return v
}

func OAuth2AuthenticationWithAccount(acct security.Account, opts ...sectest.SecurityMockOptions) oauth2.Authentication {
	var opt sectest.SecurityContextOption
	opts = append([]sectest.SecurityMockOptions{SecurityMockWithAccount(acct)}, opts...)
	sectest.MockedAuthentication(opts...)(&opt)
	return opt.Authentication.(oauth2.Authentication)
}

func SecurityMockWithAccount(acct security.Account) sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		*d = MockedSecurityDetailsTmpl
		d.Username = acct.Username()
		d.UserId = fmt.Sprintf(`%v`, acct.ID())
		d.AccountType = acct.Type()
		d.TenantId = acct.(security.AccountTenancy).DefaultDesignatedTenantId()
		d.Permissions = utils.NewStringSetFrom(d.Permissions)
		d.Permissions.Add(acct.Permissions()...)
		d.Tenants = utils.NewStringSetFrom(d.Tenants)
		d.Tenants.Add(acct.(security.AccountTenancy).DesignatedTenantIds()...)
		d.Scopes = utils.NewStringSetFrom(d.Scopes)
		d.OAuth2ResponseTypes = utils.NewStringSetFrom(d.OAuth2ResponseTypes)
		kvs := map[string]interface{}{}
		for k, v := range d.KVs {
			kvs[k] = v
		}
		d.KVs = kvs
	}
}

func AssertIDToken(g *gomega.WithT, idToken string, decoder jwt.JwtDecoder, expectAcct security.Account, expectAuth oauth2.Authentication, expectExtra ...ExpectedClaimsOption) {
	var claims IdTokenClaims
	e := decoder.DecodeWithClaims(context.Background(), idToken, &claims)
	g.Expect(e).To(Succeed(), "decoding JWT should not fail")
	AssertIDTokenClaims(g, &claims, expectAcct, expectAuth, expectExtra...)
}

func AssertIDTokenClaims(g *gomega.WithT, claims oauth2.Claims, expectAcct security.Account, expectAuth oauth2.Authentication, expectExtra ...ExpectedClaimsOption) {
	expectOpts := []ExpectedClaimsOption{
		// Basic
		ExpectClaim(oauth2.ClaimIssuer, "http://"+IssuerDomain+IssuerPath),
		ExpectClaim(oauth2.ClaimSubject, expectAcct.Username()),
		ExpectClaim(oauth2.ClaimAudience, utils.NewStringSet(ClientIDMinor)),
		ExpectClaim(oauth2.ClaimExpire, Not(BeZero())),
		ExpectClaim(oauth2.ClaimIssueAt, Not(BeZero())),
		ExpectClaim(oauth2.ClaimAuthCtxClassRef, HavePrefix("http://"+IssuerDomain+IssuerPath+"/loa-")),
		ExpectClaim(oauth2.ClaimAuthMethodRef, ContainElement("password")),
		ExpectClaim(oauth2.ClaimAuthorizedParty, Equal(ClientIDMinor)),
		ExpectClaim(oauth2.ClaimAccessTokenHash, Not(BeZero())),
		// Custom
		ExpectClaim(oauth2.ClaimUserId, expectAcct.ID()),
		ExpectClaim(oauth2.ClaimAccountType, expectAcct.Type().String()),
		ExpectClaim(oauth2.ClaimTenantId, expectAcct.(security.AccountTenancy).DefaultDesignatedTenantId()),
	}
	params := expectAuth.OAuth2Request().Parameters()
	if v, ok := params[oauth2.ParameterNonce]; ok {
		expectOpts = append(expectOpts, ExpectClaim(oauth2.ClaimNonce, v))
	}
	if _, ok := params[oauth2.ParameterMaxAge]; ok {
		expectOpts = append(expectOpts, ExpectClaim(oauth2.ClaimNonce, Not(BeZero())))
	}
	expectOpts = append(expectOpts, expectExtra...)
	AssertClaims(g, claims, NewExpectedClaims(expectOpts...))
}
