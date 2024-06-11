package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cisco-open/go-lanai/test"
	. "github.com/cisco-open/go-lanai/test/utils/gomega"
	"github.com/golang-jwt/jwt/v4"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"testing"
)

/*************************
	Test Setup
 *************************/

type MarshalExpectation struct {
	Kid      string
	Type     string
	Matchers map[string]types.GomegaMatcher
}

var (
	ExpectEC256 = ECMarshalExpectation("P-256")
	ExpectEC384 = ECMarshalExpectation("P-384")
	ExpectEC521 = ECMarshalExpectation("P-521")
	ExpectRSA   = MarshalExpectation{
		Kid:  testDefaultKid,
		Type: "RSA",
		Matchers: map[string]types.GomegaMatcher{
			"n": HaveJsonPathWithValue("$.n", ContainElements(Not(BeEmpty()))),
			"e": HaveJsonPathWithValue("$.e", ContainElements(Not(BeEmpty()))),
		},
	}
	ExpectOct = MarshalExpectation{
		Kid:  testDefaultKid,
		Type: "oct",
		Matchers: map[string]types.GomegaMatcher{
			"k": HaveJsonPathWithValue("$.k", ContainElements(Not(BeEmpty()))),
		},
	}
)

/*************************
	Test Cases
 *************************/

func TestJwkMarshaling(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodRS256, false, ExpectRSA), "RSA-KeyPair"),
		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodRS256, true, ExpectRSA), "RSA-PublicKey"),
		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodES256, false, ExpectEC256), "EC-KeyPair"),
		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodES256, true, ExpectEC256), "EC256-PublicKey"),
		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodES384, true, ExpectEC384), "EC384-PublicKey"),
		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodES512, true, ExpectEC521), "EC521-PublicKey"),
		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodHS256, false, ExpectOct), "HMAC-Secret"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestJwkMarshal(method jwt.SigningMethod, publicOnly bool, expect MarshalExpectation) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		jwk := NewTestJwk(g, method, publicOnly)
		data, e := json.Marshal(jwk)
		g.Expect(e).To(Succeed(), "marshaling JWK should not fail")
		g.Expect(data).ToNot(BeEmpty(), "marshaled JWK should not fail")
		t.Logf("JSON: %s", string(data))
		AssertMarshalResult(g, data, expect)
		parsed := AssertUnmarshal(g, data, expect)
		if !publicOnly {
			AssertJwkPair(ctx, g, parsed, jwk.(PrivateJwk))
		}
	}
}

/*************************
	Helpers
 *************************/

func NewTestJwk(g *gomega.WithT, method jwt.SigningMethod, publicOnly bool) Jwk {
	jwk, e := generateRandomJwk(method, testDefaultKid, testDefaultKid)
	g.Expect(e).To(Succeed(), "generating JWK should not fail")
	if publicOnly {
		return NewJwk(jwk.Id(), jwk.Name(), jwk.Public())
	}
	return jwk
}

func AssertMarshalResult(g *gomega.WithT, data []byte, expect MarshalExpectation) {
	g.Expect(data).To(HaveJsonPathWithValue(`$.kid`, ContainElements(expect.Kid)), `marshaled JWK should have correct 'kid'`)
	g.Expect(data).To(HaveJsonPathWithValue(`$.kty`, ContainElements(expect.Type)), `marshaled JWK should have correct 'kty'`)
	for k, v := range expect.Matchers {
		g.Expect(data).To(v, `marshaled JWK should have correct '%s'`, k)
	}
}

func AssertUnmarshal(g *gomega.WithT, data []byte, expect MarshalExpectation) Jwk {
	var parsed GenericJwk
	e := json.Unmarshal(data, &parsed)
	g.Expect(e).To(Succeed(), "unmarshalling JSON should not fail")
	g.Expect(parsed).ToNot(BeZero(), "unmarshalled JWK should not be zero")
	g.Expect(parsed.Id()).To(Equal(expect.Kid), "unmarshalled JWK should have correct kid")
	g.Expect(parsed.Name()).To(Equal(expect.Kid), "unmarshalled JWK should have correct name")
	g.Expect(parsed.Public()).ToNot(BeZero(), "unmarshalled JWK should have public key")
	return &parsed
}

func AssertJwkPair(ctx context.Context, g *gomega.WithT, pub Jwk, priv PrivateJwk) {
	privStore := &TestJwkStore{Jwk: priv}
	encoder := NewSignedJwtEncoder(SignWithJwkStore(privStore, testDefaultKid), SignWithMethod(nil))
	pubStore := &TestJwkStore{Jwk: pub}
	decoder := NewSignedJwtDecoder(VerifyWithJwkStore(pubStore, testDefaultKid), VerifyWithMethods(SupportedSigningMethods...))
	token, e := encoder.Encode(ctx, claims)
	g.Expect(e).To(Succeed(), "encode JWT with private JWK should not fail")
	decoded, e := decoder.Decode(ctx, token)
	g.Expect(e).To(Succeed(), "decode JWT with public JWK should not fail")
	g.Expect(decoded).ToNot(BeZero(), "decoded claims with public JWK should not be zero")
}

func ECMarshalExpectation(crv string) MarshalExpectation {
	return MarshalExpectation{
		Kid:  testDefaultKid,
		Type: "EC",
		Matchers: map[string]types.GomegaMatcher{
			"crv": HaveJsonPathWithValue("$.crv", ContainElements(Equal(crv))),
			"x":   HaveJsonPathWithValue("$.x", ContainElements(Not(BeEmpty()))),
			"y":   HaveJsonPathWithValue("$.y", ContainElements(Not(BeEmpty()))),
		},
	}
}

type TestJwkStore struct {
	Jwk Jwk
}

func (s TestJwkStore) LoadByKid(_ context.Context, kid string) (Jwk, error) {
	if s.Jwk.Id() == kid {
		return s.Jwk, nil
	}
	return nil, fmt.Errorf("wrong kid")
}

func (s TestJwkStore) LoadByName(_ context.Context, name string) (Jwk, error) {
	if s.Jwk.Name() == name {
		return s.Jwk, nil
	}
	return nil, fmt.Errorf("wrong key name")
}

func (s TestJwkStore) LoadAll(_ context.Context, names ...string) ([]Jwk, error) {
	panic("don't call me")
}

