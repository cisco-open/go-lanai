package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/cisco-open/go-lanai/test"
	. "github.com/cisco-open/go-lanai/test/utils/gomega"
	"github.com/golang-jwt/jwt/v4"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"os"
	"testing"
)

/*************************
	Test Setup
 *************************/

const TestDummyKid = `dummy-kid`

type MarshalExpectation struct {
	Kid          string
	Type         string
	JsonMatchers map[string]types.GomegaMatcher
	JwkMatchers  map[string]types.GomegaMatcher
}

var (
	ExpectEC256 = ECMarshalExpectation("P-256")
	ExpectEC384 = ECMarshalExpectation("P-384")
	ExpectEC521 = ECMarshalExpectation("P-521")
	ExpectRSA   = MarshalExpectation{
		Kid:  TestDummyKid,
		Type: "RSA",
		JsonMatchers: map[string]types.GomegaMatcher{
			"n": HaveJsonPathWithValue("$.n", ContainElements(Not(BeEmpty()))),
			"e": HaveJsonPathWithValue("$.e", ContainElements(Not(BeEmpty()))),
		},
		JwkMatchers: map[string]types.GomegaMatcher{
			"public key": HaveField("Public()", BeAssignableToTypeOf(&rsa.PublicKey{})),
		},
	}
	ExpectOct = MarshalExpectation{
		Kid:  TestDummyKid,
		Type: "oct",
		JsonMatchers: map[string]types.GomegaMatcher{
			"k": HaveJsonPathWithValue("$.k", ContainElements(Not(BeEmpty()))),
		},
		JwkMatchers: map[string]types.GomegaMatcher{
			"public key": HaveField("Public()", BeAssignableToTypeOf([]byte{})),
		},
	}
	ExpectOKP = MarshalExpectation{
		Kid:  TestDummyKid,
		Type: "OKP",
		JsonMatchers: map[string]types.GomegaMatcher{
			"x": HaveJsonPathWithValue("$.x", ContainElements(Not(BeEmpty()))),
		},
		JwkMatchers: map[string]types.GomegaMatcher{
			"public key": HaveField("Public()", BeAssignableToTypeOf(ed25519.PublicKey{})),
		},
	}
)

const (
	TokenES256 = `eyJhbGciOiJFUzI1NiIsImtpZCI6ImR1bW15LWtpZCIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0ZXN0ZXIiLCJjbGllbnRfaWQiOiJkdW1teS1jbGllbnQiLCJpYXQiOjE3MTY1NzYwMDMsImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODkwMC9hdXRoIiwianRpIjoiZTA3ZGY5NTUtMjhmNC00ZmUwLThiZmQtNTRkNGRmNGE2NTliIiwibmJmIjoxNzE2NTc2MDAzLCJzdWIiOiJteS11c2VyIn0.vkbw2SViTkLFRI8H4kF12qNNggXkHumSOoHWldg2tAU2-yl9iTDR8JZc-n0yz20Z0tFhp-BDgiNQ68pJxVhdMg`
	TokenRS256 = `eyJhbGciOiJSUzI1NiIsImtpZCI6ImR1bW15LWtpZCIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0ZXN0ZXIiLCJjbGllbnRfaWQiOiJkdW1teS1jbGllbnQiLCJpYXQiOjE3MTY1NzYwMDMsImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODkwMC9hdXRoIiwianRpIjoiZTA3ZGY5NTUtMjhmNC00ZmUwLThiZmQtNTRkNGRmNGE2NTliIiwibmJmIjoxNzE2NTc2MDAzLCJzdWIiOiJteS11c2VyIn0.Z--67tOijHME_F1UhMLTtTbJ6T30BNY6F-OxDL4N1Sg4_GhffNjONL8Ar8iji_IAds_DC961jYoppX9RF9BDkr33kvZsfXMAedBGjccRRn2o93-4v-OwF3aDgJxqffjhb9M2NGr5I7MLPwAiwaqW8-IkkwOeoBTwCQa7BQ5nxeptXGQ8gT0vKMAj6GnVH-2QVpBFVM5CXbJDQx7LNOzKpMYxW5LTQmrJlZamtSxyvYsp4pMh1hSe56XcPEFbKQ84ugHk5YaedGzmSAkwxv_vz-_X1Hk2T7S2aDYZ9Pey2FK8jxYOTq9vGKpQX6zggf0Qz4y7FOpMWZ8Xzw851z2S8A`
	TokenHS256 = `eyJhbGciOiJIUzI1NiIsImtpZCI6ImR1bW15LWtpZCIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0ZXN0ZXIiLCJjbGllbnRfaWQiOiJkdW1teS1jbGllbnQiLCJpYXQiOjE3MTY1NzYwMDMsImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODkwMC9hdXRoIiwianRpIjoiZTA3ZGY5NTUtMjhmNC00ZmUwLThiZmQtNTRkNGRmNGE2NTliIiwibmJmIjoxNzE2NTc2MDAzLCJzdWIiOiJteS11c2VyIn0.jwHlGBiJC_8LbA-KAfaIodb6spn_DjYb8AvAFDU53eI`
	TokenEdDSA = ``
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

		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodEdDSA, false, ExpectOKP), "ED25519-KeyPair"),
		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodEdDSA, true, ExpectOKP), "ED25519-PublicKey"),

		test.GomegaSubTest(SubTestJwkMarshal(jwt.SigningMethodHS256, false, ExpectOct), "HMAC-Secret"),
	)
}

func TestJwkUnmarshalling(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestJwkUnmarshal(jwt.SigningMethodRS256, TokenRS256, ExpectRSA), "RSA-RS256"),
		test.GomegaSubTest(SubTestJwkUnmarshal(jwt.SigningMethodES256, TokenES256, ExpectEC256), "EC-ES256"),
		test.GomegaSubTest(SubTestJwkUnmarshal(jwt.SigningMethodHS256, TokenHS256, ExpectOct), "HMAC-HS256"),
		//test.GomegaSubTest(SubTestJwkUnmarshal(jwt.SigningMethodEdDSA, TokenEdDSA, ExpectOKP), "Ed25519-EdDSA"),
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

		var parsed GenericJwk
		e = json.Unmarshal(data, &parsed)
		g.Expect(e).To(Succeed(), "unmarshalling JSON should not fail")
		AssertUnmarshalResult(g, &parsed, expect)

		if !publicOnly {
			AssertJwkPair(ctx, g, &parsed, jwk.(PrivateJwk))
		}
	}
}

func SubTestJwkUnmarshal(method jwt.SigningMethod, jwtVal string, expect MarshalExpectation) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		filename := fmt.Sprintf(`testdata/jwk-%s.json`, method.Alg())
		data, e := os.ReadFile(filename)
		g.Expect(e).To(Succeed(), "read JWK JSON file should not fail")

		jwk, e := ParseJwk(data)
		g.Expect(e).To(Succeed(), "parse JWK from JSON should not fail")
		AssertUnmarshalResult(g, jwk, expect)
		AssertJwt(ctx, g, jwk, jwtVal)
	}
}

/*************************
	Helpers
 *************************/

func NewTestJwk(g *gomega.WithT, method jwt.SigningMethod, publicOnly bool) Jwk {
	jwk, e := generateRandomJwk(method, TestDummyKid, TestDummyKid)
	g.Expect(e).To(Succeed(), "generating JWK should not fail")
	if publicOnly {
		return NewJwk(jwk.Id(), jwk.Name(), jwk.Public())
	}
	return jwk
}

func AssertMarshalResult(g *gomega.WithT, data []byte, expect MarshalExpectation) {
	g.Expect(data).To(HaveJsonPathWithValue(`$.kid`, ContainElements(expect.Kid)), `marshaled JWK should have correct 'kid'`)
	g.Expect(data).To(HaveJsonPathWithValue(`$.kty`, ContainElements(expect.Type)), `marshaled JWK should have correct 'kty'`)
	for k, v := range expect.JsonMatchers {
		g.Expect(data).To(v, `marshaled JWK should have correct '%s'`, k)
	}
}

func AssertUnmarshalResult(g *gomega.WithT, parsed Jwk, expect MarshalExpectation) {
	g.Expect(parsed).ToNot(BeZero(), "unmarshalled JWK should not be zero")
	g.Expect(parsed.Id()).To(Equal(expect.Kid), "unmarshalled JWK should have correct kid")
	g.Expect(parsed.Name()).To(Equal(expect.Kid), "unmarshalled JWK should have correct name")
	for k, v := range expect.JwkMatchers {
		g.Expect(parsed).To(v, "unmarshalled JWK should have correct %s", k)
	}
}

func AssertJwkPair(ctx context.Context, g *gomega.WithT, pub Jwk, priv PrivateJwk) {
	privStore := &TestJwkStore{Jwk: priv}
	encoder := NewSignedJwtEncoder(SignWithJwkStore(privStore, TestDummyKid), SignWithMethod(nil))
	pubStore := &TestJwkStore{Jwk: pub}
	decoder := NewSignedJwtDecoder(VerifyWithJwkStore(pubStore, TestDummyKid))
	token, e := encoder.Encode(ctx, claims)
	g.Expect(e).To(Succeed(), "encode JWT with private JWK should not fail")
	decoded, e := decoder.Decode(ctx, token)
	g.Expect(e).To(Succeed(), "decode JWT with public JWK should not fail")
	g.Expect(decoded).ToNot(BeZero(), "decoded claims with public JWK should not be zero")
}

func AssertJwt(ctx context.Context, g *gomega.WithT, pub Jwk, token string) {
	pubStore := &TestJwkStore{Jwk: pub}
	decoder := NewSignedJwtDecoder(VerifyWithJwkStore(pubStore, TestDummyKid))
	decoded, e := decoder.Decode(ctx, token)
	g.Expect(e).To(Succeed(), "decode JWT with public JWK should not fail")
	g.Expect(decoded).ToNot(BeZero(), "decoded claims with public JWK should not be zero")
}

func ECMarshalExpectation(crv string) MarshalExpectation {
	return MarshalExpectation{
		Kid:  TestDummyKid,
		Type: "EC",
		JsonMatchers: map[string]types.GomegaMatcher{
			"crv": HaveJsonPathWithValue("$.crv", ContainElements(Equal(crv))),
			"x":   HaveJsonPathWithValue("$.x", ContainElements(Not(BeEmpty()))),
			"y":   HaveJsonPathWithValue("$.y", ContainElements(Not(BeEmpty()))),
		},
		JwkMatchers: map[string]types.GomegaMatcher{
			"public key": HaveField("Public()", BeAssignableToTypeOf(&ecdsa.PublicKey{})),
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
