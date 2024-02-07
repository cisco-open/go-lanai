package misc_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/misc"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/openid"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Setup Test
 *************************/

/*************************
	Test
 *************************/

type WellKnownDI struct {
	fx.In
	Issuer security.Issuer
}

func TestWellKnownEndpoint(t *testing.T) {
	var di WellKnownDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithFxOptions(
			fx.Provide(
				BindMockingProperties,
				NewTestIssuer, NewTestClientStore,
			),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestOpenIDConfigWithoutExtra(&di), "OpenIDConfigWithoutExtra"),
		test.GomegaSubTest(SubTestOpenIDConfigWithExtra(&di), "OpenIDConfigWithExtra"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestOpenIDConfigWithoutExtra(di *WellKnownDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const username = TestUser1
		var req *http.Request
		var resp *openid.OPMetadata
		var e error
		idpManager := sectest.NewMockedIDPManager(func(opt *sectest.IdpManagerMockOption) {
			opt.PasswdIDPDomain = IssuerDomain
		})
		endpoint := misc.NewWellKnownEndpoint(di.Issuer, idpManager, nil)
		resp, e = endpoint.OpenIDConfig(ctx, req)
		g.Expect(e).To(Succeed(), "OpenIDConfig should fail without authentication")
		AssertOpenIDConfigClaims(g, resp)
	}
}

func SubTestOpenIDConfigWithExtra(di *WellKnownDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *openid.OPMetadata
		var e error
		idpManager := sectest.NewMockedIDPManager(func(opt *sectest.IdpManagerMockOption) {
			opt.PasswdIDPDomain = IssuerDomain
		})
		endpoint := misc.NewWellKnownEndpoint(di.Issuer, idpManager, map[string]interface{}{
			openid.OPMetadataAuthEndpoint:       "/authorize",
			openid.OPMetadataTokenEndpoint:      "/token",
			openid.OPMetadataUserInfoEndpoint:   "/userinfo",
			openid.OPMetadataJwkSetURI:          "/jwks",
			openid.OPMetadataEndSessionEndpoint: "/logout",
		})

		resp, e = endpoint.OpenIDConfig(ctx, req)
		g.Expect(e).To(Succeed(), "OpenIDConfig should fail without authentication")
		AssertOpenIDConfigClaims(g, resp,
			ExpectClaim(openid.OPMetadataAuthEndpoint, FullURL("/authorize")),
			ExpectClaim(openid.OPMetadataTokenEndpoint, FullURL("/token")),
			ExpectClaim(openid.OPMetadataUserInfoEndpoint, FullURL("/userinfo")),
			ExpectClaim(openid.OPMetadataJwkSetURI, FullURL("/jwks")),
		)
	}
}

/*************************
	Helpers
 *************************/

func ACRValue(lvl int) string {
	return fmt.Sprintf("http://%s%s/loa-%d", IssuerDomain, IssuerPath, lvl)
}

func FullURL(path string) string {
	return fmt.Sprintf("http://%s%s%s", IssuerDomain, IssuerPath, path)
}

func AssertOpenIDConfigClaims(g *gomega.WithT, claims oauth2.Claims, expectExtra ...ExpectedClaimsOption) {
	expectOpts := []ExpectedClaimsOption{
		ExpectClaim(openid.OPMetadataIssuer, "http://"+IssuerDomain+IssuerPath),
		ExpectClaim(openid.OPMetadataGrantTypes, HaveLen(8)),
		ExpectClaim(openid.OPMetadataScopes, HaveLen(9)),
		ExpectClaim(openid.OPMetadataResponseTypes, HaveKey("code")),
		ExpectClaim(openid.OPMetadataACRValues, HaveKey(Or(Equal(ACRValue(1)), Equal(ACRValue(2)), Equal(ACRValue(3))))),
		ExpectClaim(openid.OPMetadataSubjectTypes, HaveKey("public")),
		ExpectClaim(openid.OPMetadataIdTokenJwsAlg, HaveKey("RS256")),
		ExpectClaim(openid.OPMetadataClaims, Not(BeEmpty())),
	}
	expectOpts = append(expectOpts, expectExtra...)
	AssertClaims(g, claims, NewExpectedClaims(expectOpts...))
}
