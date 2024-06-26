package misc_test

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth/misc"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/rest"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
	. "github.com/cisco-open/go-lanai/test/utils/gomega"
	"github.com/cisco-open/go-lanai/test/webtest"
	gojwt "github.com/golang-jwt/jwt/v4"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"go.uber.org/fx"
	"io"
	"net/http"
	"reflect"
	"testing"
)

/*************************
	Setup Test
 *************************/

type ExpectedJwk struct {
	PathPrefix   string
	Alg          gojwt.SigningMethod
	Kid          string
	Type         string
	JsonFields map[string]types.GomegaMatcher
}

var ExpectedJwks = []ExpectedJwk{
	{ PathPrefix: "/rsa", Alg: gojwt.SigningMethodRS256, Kid: TestJwtKID1, Type: "RSA", JsonFields: map[string]types.GomegaMatcher{
		"e": Not(BeEmpty()),
		"n": Not(BeEmpty()),
	}},
	{ PathPrefix: "/ec", Alg: gojwt.SigningMethodES256, Kid: TestJwtKID1, Type: "EC", JsonFields: map[string]types.GomegaMatcher{
		"crv": Not(BeEmpty()),
		"x": Not(BeEmpty()),
		"y": Not(BeEmpty()),

	}},
	{ PathPrefix: "/ed", Alg: gojwt.SigningMethodEdDSA, Kid: TestJwtKID1, Type: "OKP", JsonFields: map[string]types.GomegaMatcher{
		"x": Not(BeEmpty()),
	}},
	{ PathPrefix: "/mac", Alg: gojwt.SigningMethodHS256, Kid: TestJwtKID1, Type: "oct", JsonFields: map[string]types.GomegaMatcher{
		"k": Not(BeEmpty()),
	}},
}

type JwksDI struct {
	fx.In
	WebReg   *web.Registrar
}

func RegisterJwksEndpoint(di JwksDI) {
	for i := range ExpectedJwks {
		store := jwt.NewStaticJwkStoreWithOptions(func(s *jwt.StaticJwkStore) {
			s.KIDs = []string{TestJwtKID1, TestJwtKID2}
			s.SigningMethod = ExpectedJwks[i].Alg
		})
		ep := misc.NewJwkSetEndpoint(store)
		di.WebReg.MustRegister(rest.Get(ExpectedJwks[i].PathPrefix + "/jwks").EndpointFunc(ep.JwkSet).Build())
		di.WebReg.MustRegister(rest.Get(ExpectedJwks[i].PathPrefix + "/jwks/:kid").EndpointFunc(ep.JwkByKid).Build())
	}

}

/*************************
	Test
 *************************/

func TestJwkSetEndpoint(t *testing.T) {
	var di JwksDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithFxOptions(
			fx.Provide(sectest.BindMockingProperties),
			fx.Invoke(RegisterJwksEndpoint),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestJwkSetWithType(), "JwkSet"),
		test.GomegaSubTest(SubTestJwkByKidWithType(), "JwkByKid"),
		test.GomegaSubTest(SubTestJwkByKidFailure(), "JwkByKidFailure"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestJwkSetWithType() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		tests := make([]test.Options, len(ExpectedJwks))
		for i := range ExpectedJwks {
			tests[i] = test.GomegaSubTest(SubTestJwkSet(ExpectedJwks[i]), "JwkSet"+ExpectedJwks[i].Type)
		}
		test.RunTest(ctx, t, tests...)
	}
}

func SubTestJwkByKidWithType() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		tests := make([]test.Options, len(ExpectedJwks))
		for i := range ExpectedJwks {
			tests[i] = test.GomegaSubTest(SubTestJwkByKidSuccess(ExpectedJwks[i]), "JwkByKid"+ExpectedJwks[i].Type)
		}
		test.RunTest(ctx, t, tests...)
	}
}

func SubTestJwkByKidFailure() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		req := webtest.NewRequest(ctx, http.MethodGet, "/rsa/jwks/not-exist", nil)
		resp := webtest.MustExec(ctx, req).Response
		AssertJwkResponse(g, resp, ExpectedJwk{})
	}
}

func SubTestJwkSet(expect ExpectedJwk) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		req := webtest.NewRequest(ctx, http.MethodGet, expect.PathPrefix + "/jwks", nil)
		resp := webtest.MustExec(ctx, req).Response
		AssertJwkSetResponse(g, resp, expect)
	}
}

func SubTestJwkByKidSuccess(expect ExpectedJwk) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		req := webtest.NewRequest(ctx, http.MethodGet, expect.PathPrefix + "/jwks/" + expect.Kid, nil)
		resp := webtest.MustExec(ctx, req).Response
		AssertJwkResponse(g, resp, expect)
	}
}

/*************************
	Helpers
 *************************/

func AssertJwkSetResponse(g *gomega.WithT, resp *http.Response, expect ExpectedJwk) {
	g.Expect(resp).ToNot(BeNil(), "jwks endpoint should have response")
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "jwks endpoint should not fail")
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), "read response body should not fail")
	g.Expect(body).To(JwkSetJsonPathMatcher("kid", expect.Kid, false), "body should have correct kid")
	g.Expect(body).To(JwkSetJsonPathMatcher("kty", expect.Type, true), "body should have correct kty")
	for k, v := range expect.JsonFields {
		g.Expect(body).To(JwkSetJsonPathMatcher(k, v, true), "body should have correct %s", k)
	}
}

func AssertJwkResponse(g *gomega.WithT, resp *http.Response, expect ExpectedJwk) {
	g.Expect(resp).ToNot(BeNil(), "jwks/<kid> endpoint should have response")
	if reflect.ValueOf(expect).IsZero() {
		g.Expect(resp.StatusCode).To(Equal(http.StatusNotFound), "jwks/<kid> endpoint should fail")
		return
	}
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "jwks/<kid> endpoint should not fail")
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), "read response body should not fail")
	g.Expect(body).To(JwkJsonPathMatcher("kid", expect.Kid), "body should have correct kid")
	g.Expect(body).To(JwkJsonPathMatcher("kty", expect.Type), "body should have correct kty")
	for k, v := range expect.JsonFields {
		g.Expect(body).To(JwkJsonPathMatcher(k, v), "body should have correct %s", k)
	}
}

func JwkSetJsonPathMatcher(jwkField string, value interface{}, all bool) types.GomegaMatcher {
	jsonPath := fmt.Sprintf(`$.keys[*].%s`, jwkField)
	if all {
		return HaveJsonPathWithValue(jsonPath, HaveEach(value))
	}
	return HaveJsonPathWithValue(jsonPath, ContainElements(value))
}

func JwkJsonPathMatcher(jwkField string, value interface{}) types.GomegaMatcher {
	jsonPath := fmt.Sprintf(`$.%s`, jwkField)
	return HaveJsonPathWithValue(jsonPath, HaveEach(value))
}