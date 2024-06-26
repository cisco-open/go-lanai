package misc_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"go.uber.org/fx"
	"time"
)

/*************************
	Mocking
 *************************/

const (
	IssuerDomain  = `misc.test`
	IssuerPath    = `/auth`
	TestUser1     = `test-user-1`
	TestUser2     = `test-user-2`
	TestTenantID  = `id-root`
	ClientIDSuper = `super-client`
	ClientIDMinor = `minor-client`
	TestJwtKID1 = `test-key-1`
	TestJwtKID2 = `test-key-2`
)

type AuthDI struct {
	fx.In
	ClientStore oauth2.OAuth2ClientStore
	Mocking     sectest.MockingProperties
}

func NewTestIssuer() security.Issuer {
	return security.NewIssuer(func(details *security.DefaultIssuerDetails) {
		*details = security.DefaultIssuerDetails{
			Protocol:    "http",
			Domain:      IssuerDomain,
			Port:        80,
			ContextPath: IssuerPath,
			IncludePort: false,
		}
	})
}

func NewTestTokenStoreReader(props sectest.MockingProperties) oauth2.TokenStoreReader {
	return sectest.NewMockedTokenStoreReader(props.Accounts.MapValues(), props.Tenants.MapValues())
}

func NewTestClientStore(props sectest.MockingProperties) oauth2.OAuth2ClientStore {
	return sectest.NewMockedClientStore(props.Clients.Values()...)
}

func NewTestAccountStore(props sectest.MockingProperties) security.AccountStore {
	return sectest.NewMockedAccountStore(props.Accounts.Values())
}

func NewJwtEncoder(jwks jwt.JwkStore) jwt.JwtEncoder {
	return jwt.NewSignedJwtEncoder(jwt.SignWithJwkStore(jwks, TestJwtKID1))
}

func NewJwtDecoder(jwks jwt.JwkStore) jwt.JwtDecoder {
	return jwt.NewSignedJwtDecoder(jwt.VerifyWithJwkStore(jwks, TestJwtKID1))
}

func NewJwkStore() jwt.JwkStore {
	return jwt.NewSingleJwkStoreWithOptions(func(s *jwt.SingleJwkStore) {
		s.Kid = TestJwtKID1
	})
}

type MockedClientAuth struct {
	Client oauth2.OAuth2Client
}

func MockedTokenValue(username, tenantId string, exp time.Time, scopes ...string) string {
	now := time.Now()
	t := sectest.MockedToken{
		MockedTokenInfo: sectest.MockedTokenInfo{
			UName:  username,
			TID:    tenantId,
			Exp:    now.Unix(),
			Iss:    now.Unix(),
			Scopes: append([]string{"read", "write"}, scopes...),
		},
		ExpTime: exp,
		IssTime: now,
	}
	text, e := t.MarshalText()
	if e != nil {
		return ""
	}
	return string(text)
}

func (a MockedClientAuth) Principal() interface{} {
	return a.Client
}

func (a MockedClientAuth) Permissions() security.Permissions {
	perms := security.Permissions{}
	for scope := range a.Client.Scopes() {
		perms[scope] = true
	}
	return perms
}

func (a MockedClientAuth) State() security.AuthenticationState {
	return security.StateAuthenticated
}

func (a MockedClientAuth) Details() interface{} {
	return nil
}

/*************************
	Common Helpers
 *************************/

func ContextWithClient(ctx context.Context, g *gomega.WithT, di *AuthDI, clientId string) context.Context {
	client, e := di.ClientStore.LoadClientByClientId(ctx, clientId)
	g.Expect(e).To(Succeed(), "client [%s] should exists", clientId)
	auth := MockedClientAuth{Client: client}
	ctx = utils.MakeMutableContext(ctx)
	security.MustSet(ctx, auth)
	return context.WithValue(ctx, oauth2.CtxKeyAuthenticatedClient, client)
}

type ExpectedClaims struct {
	KVs map[string]interface{}
}

type ExpectedClaimsOption func(expect *ExpectedClaims)

func NewExpectedClaims(opts ...ExpectedClaimsOption) *ExpectedClaims {
	expect := ExpectedClaims{
		KVs: map[string]interface{}{},
	}
	for _, fn := range opts {
		fn(&expect)
	}
	return &expect
}

func ExpectClaim(k string, v interface{}) ExpectedClaimsOption {
	return func(expect *ExpectedClaims) {
		expect.KVs[k] = v
	}
}

func AssertClaims(g *gomega.WithT, claims oauth2.Claims, expect *ExpectedClaims) {
	g.Expect(claims).ToNot(BeNil(), "claims should not be nil")
	for claim, value := range expect.KVs {
		switch v := value.(type) {
		case nil:
			g.Expect(claims.Has(claim)).To(BeFalse(), "claims should not contain [%s]", claim)
		case types.GomegaMatcher:
			g.Expect(claims.Get(claim)).To(v, "claims should have correct claim [%s]", claim)
		default:
			g.Expect(claims.Get(claim)).To(BeEquivalentTo(value), "claims should have correct claim [%s]", claim)
		}
	}
}
