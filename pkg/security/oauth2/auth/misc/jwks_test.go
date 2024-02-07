package misc_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/misc"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Setup Test
 *************************/

/*************************
	Test
 *************************/

type JwksDI struct {
	fx.In
	JwkStore jwt.JwkStore
}

func TestJwkSetEndpoint(t *testing.T) {
	var di JwksDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithFxOptions(
			fx.Provide(
				BindMockingProperties, NewJwkStore,
			),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestJwkSet(&di), "JwkSet"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestJwkSet(di *JwksDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req misc.JwkSetRequest
		var resp *misc.JwkSetResponse
		var e error
		endpoint := misc.NewJwkSetEndpoint(di.JwkStore)
		resp, e = endpoint.JwkSet(ctx, &req)
		g.Expect(e).To(Succeed(), "JwkSet should not fail without authentication")
		g.Expect(resp).ToNot(BeNil(), "response should not be nil")
		g.Expect(resp.Keys).ToNot(BeEmpty(), "response should contains keys")
		var found bool
		for _, key := range resp.Keys {
			g.Expect(key.Type).To(Equal("RSA"))
			g.Expect(key.Modulus).ToNot(BeEmpty())
			g.Expect(key.Exponent).ToNot(BeEmpty())
			if key.Id == JwtKID {
				found = true
			}
		}
		g.Expect(found).To(BeTrue(), "response should contains a key with KID = %s", JwtKID)
	}
}

/*************************
	Helpers
 *************************/

