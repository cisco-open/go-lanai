package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
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
	JwtKID    = `test-key`
	TestNonce = "please give it back"
)

const MockingPrefix = "mocking"

type MockingProperties struct {
	Accounts map[string]*sectest.MockedAccountProperties `json:"accounts"`
	Tenants  map[string]*sectest.MockedTenantProperties  `json:"tenants"`
	Clients  map[string]*sectest.MockedClientProperties  `json:"clients"`
}

func BindMockingProperties(appCtx *bootstrap.ApplicationContext) MockingProperties {
	props := MockingProperties{
		Accounts: map[string]*sectest.MockedAccountProperties{},
		Tenants:  map[string]*sectest.MockedTenantProperties{},
	}
	if e := appCtx.Config().Bind(&props, MockingPrefix); e != nil {
		panic(e)
	}
	return props
}

type AuthDI struct {
	fx.In
	ClientStore oauth2.OAuth2ClientStore
	Mocking     MockingProperties
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

func NewTestTokenStoreReader(props MockingProperties) oauth2.TokenStoreReader {
	return sectest.NewMockedTokenStoreReader(props.Accounts, props.Tenants)
}

func NewTestClientStore(props MockingProperties) oauth2.OAuth2ClientStore {
	clients := make([]*sectest.MockedClientProperties, 0, len(props.Clients))
	for _, c := range props.Clients {
		clients = append(clients, c)
	}
	return sectest.NewMockedClientStore(clients...)
}

func NewTestAccountStore(props MockingProperties) security.AccountStore {
	accts := make([]*sectest.MockedAccountProperties, 0, len(props.Accounts))
	for _, acct := range props.Accounts {
		accts = append(accts, acct)
	}

	tenants := make([]*sectest.MockedTenantProperties, 0, len(props.Tenants))
	for _, t := range props.Tenants {
		tenants = append(tenants, t)
	}
	return sectest.NewMockedAccountStore(accts, tenants)
}

func NewJwtEncoder(jwks jwt.JwkStore) jwt.JwtEncoder {
	return jwt.NewRS256JwtEncoder(jwks, JwtKID)
}

func NewJwtDecoder(jwks jwt.JwkStore) jwt.JwtDecoder {
	return jwt.NewRS256JwtDecoder(jwks, JwtKID)
}

func NewJwkStore() jwt.JwkStore {
	return jwt.NewSingleJwkStore(JwtKID)
}

type MockedClientAuth struct {
	Client oauth2.OAuth2Client
}

func MockedTokenValue(username, tenantId string, exp time.Time, scopes ...string) string {
	now := time.Now()
	t := sectest.MockedToken{
		MockedTokenInfo: sectest.MockedTokenInfo{
			UName: username,
			TID:   tenantId,
			Exp:   now.Unix(),
			Iss:   now.Unix(),
		},
		ExpTime:      exp,
		IssTime:      now,
		MockedScopes: append([]string{"read", "write"}, scopes...),
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
	ctx = context.WithValue(ctx, security.ContextKeySecurity, auth)
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
