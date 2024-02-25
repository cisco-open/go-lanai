package misc_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/misc"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

/*************************
	Setup Test
 *************************/

/*************************
	Test
 *************************/

type CheckTokenDI struct {
	fx.In
	AuthDI
	Endpoint         *misc.CheckTokenEndpoint
	TokenStoreReader oauth2.TokenStoreReader
}

func TestCheckTokenEndpoint(t *testing.T) {
	var di CheckTokenDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithFxOptions(
			fx.Provide(
				sectest.BindMockingProperties,
				NewTestIssuer,
				NewTestTokenStoreReader,
				NewTestClientStore,
			),
			fx.Provide(misc.NewCheckTokenEndpoint),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestValidTokenWithDetails(&di), "ValidTokenWithDetails"),
		test.GomegaSubTest(SubTestValidTokenWithoutDetails(&di), "ValidTokenWithoutDetails"),
		test.GomegaSubTest(SubTestValidTokenWithoutDetailsScope(&di), "ValidTokenWithoutDetailsScope"),
		test.GomegaSubTest(SubTestExpiredToken(&di), "ExpiredToken"),
		test.GomegaSubTest(SubTestRevokedToken(&di), "RevokedToken"),
		test.GomegaSubTest(SubTestRefreshToken(&di), "RefreshToken"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestValidTokenWithDetails(di *CheckTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *misc.CheckTokenRequest
		var resp *misc.CheckTokenClaims
		var e error
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDSuper)
		req = &misc.CheckTokenRequest{
			Token:     MockedTokenValue(TestUser1, TestTenantID, time.Now().Add(time.Minute)),
			NoDetails: false,
		}
		resp, e = di.Endpoint.CheckToken(ctx, req)
		g.Expect(e).To(Succeed(), "CheckToken with valid token should not fail")
		AssertCheckTokenClaims(g, resp, true, true)
	}
}

func SubTestValidTokenWithoutDetails(di *CheckTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *misc.CheckTokenRequest
		var resp *misc.CheckTokenClaims
		var e error
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDSuper)
		req = &misc.CheckTokenRequest{
			Token:     MockedTokenValue(TestUser1, TestTenantID, time.Now().Add(time.Minute)),
			NoDetails: true,
		}
		resp, e = di.Endpoint.CheckToken(ctx, req)
		g.Expect(e).To(Succeed(), "CheckToken with valid token should not fail")
		AssertCheckTokenClaims(g, resp, true, false)
	}
}

func SubTestValidTokenWithoutDetailsScope(di *CheckTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *misc.CheckTokenRequest
		var resp *misc.CheckTokenClaims
		var e error
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDMinor)
		req = &misc.CheckTokenRequest{
			Token:     MockedTokenValue(TestUser1, TestTenantID, time.Now().Add(time.Minute)),
			NoDetails: false,
		}
		resp, e = di.Endpoint.CheckToken(ctx, req)
		g.Expect(e).To(Succeed(), "CheckToken with valid token should not fail")
		AssertCheckTokenClaims(g, resp, true, false)
	}
}

func SubTestExpiredToken(di *CheckTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *misc.CheckTokenRequest
		var resp *misc.CheckTokenClaims
		var e error
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDMinor)
		req = &misc.CheckTokenRequest{
			Token:     MockedTokenValue(TestUser1, TestTenantID, time.Now().Add(-time.Minute)),
			NoDetails: false,
		}
		resp, e = di.Endpoint.CheckToken(ctx, req)
		g.Expect(e).To(Succeed(), "CheckToken with expired token should not fail")
		AssertCheckTokenClaims(g, resp, false, false)
	}
}

func SubTestRevokedToken(di *CheckTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *misc.CheckTokenRequest
		var resp *misc.CheckTokenClaims
		var e error
		token := MockedTokenValue(TestUser1, TestTenantID, time.Now().Add(time.Minute))
		di.TokenStoreReader.(sectest.MockedTokenRevoker).Revoke(token)
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDMinor)
		req = &misc.CheckTokenRequest{
			Token:     token,
			NoDetails: false,
		}
		resp, e = di.Endpoint.CheckToken(ctx, req)
		g.Expect(e).To(Succeed(), "CheckToken with revoked token should not fail")
		AssertCheckTokenClaims(g, resp, false, false)
	}
}

func SubTestRefreshToken(di *CheckTokenDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *misc.CheckTokenRequest
		var resp *misc.CheckTokenClaims
		var e error
		token := MockedTokenValue(TestUser1, TestTenantID, time.Now().Add(time.Minute))
		di.TokenStoreReader.(sectest.MockedTokenRevoker).Revoke(token)
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDMinor)
		req = &misc.CheckTokenRequest{
			Token:     token,
			Hint: "refresh_token",
			NoDetails: false,
		}
		resp, e = di.Endpoint.CheckToken(ctx, req)
		g.Expect(e).To(HaveOccurred(), "CheckToken with refresh token should fail")
		g.Expect(resp).To(BeNil())
	}
}

/*************************
	Helpers
 *************************/

func AssertCheckTokenClaims(g *gomega.WithT, resp *misc.CheckTokenClaims, expectActive bool, expectDetails bool) {
	expectOpts := make([]ExpectedClaimsOption, 1, 5)
	if expectActive {
		expectOpts[0] = ExpectClaim(oauth2.ClaimActive, &utils.TRUE)
	} else {
		expectOpts[0] = ExpectClaim(oauth2.ClaimActive, &utils.FALSE)
	}
	if !expectDetails {
		expectOpts = append(expectOpts,
			ExpectClaim(oauth2.ClaimUsername, nil),
			ExpectClaim(oauth2.ClaimTenantId, nil),
			ExpectClaim(oauth2.ClaimUserId, nil),
		)
	} else {
		expectOpts = append(expectOpts,
			ExpectClaim(oauth2.ClaimUsername, Not(BeEmpty())),
			ExpectClaim(oauth2.ClaimTenantId, Not(BeEmpty())),
			ExpectClaim(oauth2.ClaimUserId, Not(BeEmpty())),
		)
	}
	AssertClaims(g, resp, NewExpectedClaims(expectOpts...))
}
