package openid

import (
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
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
	JwtKID    = `test-key`
	TestNonce = "please give it back"
)

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

func NewTestAccountStore(props sectest.MockingProperties) security.AccountStore {
	return sectest.NewMockedAccountStore(props.Accounts.Values())
}

func NewJwtEncoder(jwks jwt.JwkStore) jwt.JwtEncoder {
	return jwt.NewSignedJwtEncoder(jwt.SignWithJwkStore(jwks, JwtKID))
}

func NewJwtDecoder(jwks jwt.JwkStore) jwt.JwtDecoder {
	return jwt.NewSignedJwtDecoder(jwt.VerifyWithJwkStore(jwks, JwtKID))
}

func NewJwkStore() jwt.JwkStore {
	return jwt.NewSingleJwkStoreWithOptions(func(s *jwt.SingleJwkStore) {
		s.Kid = JwtKID
	})
}

/*************************
	Common Helpers
 *************************/

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
