package jwt

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"reflect"
	"testing"
	"time"
)

const (
	testDefaultKid = "default"
)

var claims = oauth2.MapClaims {
	"aud": []interface{}{"target"},
	"exp": time.Now().Add(24 * time.Hour).Unix(),
	"jti": uuid.New().String(),
	"iat": time.Now().Unix(),
	"nbf": time.Now().Unix(),
	"iss": "sandbox",
	"sub": "user",
}

type SubTest func(*testing.T)

/*************************
	Test Cases
 *************************/
func TestJwtWithKid(t *testing.T) {

	kids := []string{"kid1", "kid2", "kid3"}
	staticJwkStore := NewStaticJwkStore(kids...)
	enc := NewRS256JwtEncoder(staticJwkStore, testDefaultKid)

	// encoding
	value, err := enc.Encode(context.Background(), claims)
	g := NewWithT(t)
	g.Expect(err).NotTo(HaveOccurred(), "Encode shouldn't returns error")
	g.Expect(value).NotTo(BeZero(), "Encoded jwt shouldn't be empty")
	//TODO more test cases

	t.Logf("JWT: %s", value)

	// decode, happy path
	t.Run("JwtDecodeSuccessWithSameKey",
		SubTestJwtDecodeSuccessWithSameKey(value, staticJwkStore))
	t.Run("JwtDecodeSuccessWithRotatedKey",
		SubTestJwtDecodeSuccessWithRotatedKey(value, staticJwkStore))
	t.Run("JwtDecodeSuccessWithCustomClaims",
		SubTestJwtDecodeSuccessWithCustomClaims(value, staticJwkStore))

	// decode, not so happey, kid exists, but not same key
	t.Run("JwtDecodeFailedWithWrongKey",
		SubTestJwtDecodeFailedWithWrongKey(value, kids[0]))
	t.Run("JwtDecodeFailedWithNonExistingKey",
		SubTestJwtDecodeFailedWithNonExistingKey(value))
}

func TestJwtWithoutKid(t *testing.T) {
	// Note, when using default "kid" defined in Encoder, "kid" field is omitted in the JWT
	nonRotatingJwkStore := NewSingleJwkStore(testDefaultKid)
	enc := NewRS256JwtEncoder(nonRotatingJwkStore, testDefaultKid)

	// encoding
	value, err := enc.Encode(context.Background(), claims)
	g := NewWithT(t)
	g.Expect(err).NotTo(HaveOccurred(), "Encode shouldn't returns error")
	g.Expect(value).NotTo(BeZero(), "Encoded jwt shouldn't be empty")
	//TODO more test cases

	t.Logf("JWT: %s", value)

	// decode, happy path
	t.Run("JwtDecodeSuccessWithSameKey",
		SubTestJwtDecodeSuccessWithSameKey(value, nonRotatingJwkStore))
	t.Run("JwtDecodeSuccessWithCustomClaims",
		SubTestJwtDecodeSuccessWithCustomClaims(value, nonRotatingJwkStore))

	// decode, not so happey, kid exists, but not same key
	t.Run("JwtDecodeFailedWithWrongKey",
		SubTestJwtDecodeFailedWithWrongKey(value, testDefaultKid))
	t.Run("JwtDecodeFailedWithNonExistingKey",
		SubTestJwtDecodeFailedWithNonExistingKey(value))
}

/*************************
	Sub-Test Cases
 *************************/
func SubTestJwtDecodeSuccessWithSameKey(jwtVal string, jwkStore JwkStore) SubTest {
	return func(t *testing.T) {
		dec := NewRS256JwtDecoder(jwkStore, testDefaultKid)
		parsed, err := dec.Decode(context.Background(), jwtVal)

		g := NewWithT(t)
		assertDecodeResult(g, parsed, err)
		assertMapClaims(g, claims, parsed)
	}
}

func SubTestJwtDecodeSuccessWithRotatedKey(jwtVal string, jwkStore JwkRotator) SubTest {
	return func(t *testing.T) {
		if err := jwkStore.Rotate(context.Background(), ""); err != nil {
			t.Errorf("StaticJwkStore key roation should not have error, but got %v", err)
		}
		dec := NewRS256JwtDecoder(jwkStore, testDefaultKid)
		parsed, err := dec.Decode(context.Background(), jwtVal)

		g := NewWithT(t)
		assertDecodeResult(g, parsed, err)
		assertMapClaims(g, claims, parsed)
	}
}

func SubTestJwtDecodeSuccessWithCustomClaims(jwtVal string, jwkStore JwkStore) SubTest {
	return func(t *testing.T) {
		g := NewWithT(t)
		dec := NewRS256JwtDecoder(jwkStore, testDefaultKid)

		custom := customClaims{}
		err := dec.DecodeWithClaims(context.Background(), jwtVal, &custom)
		assertDecodeResult(g, custom, err)
		assertCustomClaims(g, claims, custom)

		compatible := customCompatibleClaims{}
		err = dec.DecodeWithClaims(context.Background(), jwtVal, &compatible)
		assertDecodeResult(g, custom, err)
		assertCustomClaims(g, claims, compatible.customClaims)
	}
}

func SubTestJwtDecodeFailedWithWrongKey(jwtVal string, kid string) SubTest {
	return func(t *testing.T) {

		store := NewSingleJwkStore(kid)
		dec := NewRS256JwtDecoder(store, testDefaultKid)
		_, err := dec.Decode(context.Background(), jwtVal)

		g := NewWithT(t)
		g.Expect(err).
			NotTo(Succeed(), "decode with different JWK should return validation error")
	}
}

func SubTestJwtDecodeFailedWithNonExistingKey(jwtVal string) SubTest {
	return func(t *testing.T) {
		store := NewSingleJwkStore("whatever")
		dec := NewRS256JwtDecoder(store, testDefaultKid)
		_, err := dec.Decode(context.Background(), jwtVal)

		g := NewWithT(t)
		g.Expect(err).
			NotTo(Succeed(), "decode with non-existing JWK should return validation error")
	}
}

/*************************
	Helpers
 *************************/
func assertDecodeResult(g *WithT, decoded oauth2.Claims, err error) {
	g.Expect(err).NotTo(HaveOccurred(), "Decode should not return error.")
	g.Expect(decoded).NotTo(BeNil(), "Decode should return non-nil claims")
}

func assertMapClaims(g *WithT, expected oauth2.MapClaims, decoded oauth2.Claims) {

	g.Expect(decoded).To(BeAssignableToTypeOf(oauth2.MapClaims{}), "MapClaims is expected")
	actual := decoded.(oauth2.MapClaims)

	g.Expect(len(actual)).To(Equal(len(expected)), "actual MapClaims should have same size")
	for k,v := range actual {
		g.Expect(v).To(BeEquivalentTo(expected[k]), "actual MapClaims should have same [%s]", k)
	}
}

func assertCustomClaims(g *WithT, expected oauth2.MapClaims, decoded oauth2.Claims) {

	g.Expect(decoded).To(BeAssignableToTypeOf(customClaims{}), "custom claims is expected")
	actual := decoded.(customClaims)

	for k,v := range expected {
		g.Expect(actual.Get(k)).To(BeEquivalentTo(v), "actual claims should have same [%s]", k)
	}
}

// customClaims implements Claims
type customClaims struct {
	oauth2.FieldClaimsMapper
	Audiance  []interface{} `claim:"aud"`
	Expiry    int64         `claim:"exp"`
	Id        string        `claim:"jti"`
	IssueAt   int64         `claim:"iat"`
	NotBefore int64         `claim:"nbf"`
	Issuer    string        `claim:"iss"`
	Subject   string        `claim:"sub"`
}

func (c *customClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *customClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c customClaims) Get(claim string) interface{} {
	return c.value(claim).Interface()
}

func (c customClaims) Has(claim string) bool {
	return !c.value(claim).IsZero()
}

func (c customClaims) Set(claim string, value interface{}) {
	panic("we don't support this")
}

func (c customClaims) Values() map[string]interface{} {
	return map[string]interface{}{
		`aud`: c.Audiance,
		`exp`: c.Expiry,
		`jti`: c.Id,
		`iat`: c.IssueAt,
		`nbf`: c.NotBefore,
		`iss`: c.Issuer,
		`sub`: c.Subject,
	}
}

func (c customClaims) value(claim string) reflect.Value {
	switch claim {
	case "aud":
		return reflect.ValueOf(c.Audiance)
	case "exp":
		return reflect.ValueOf(c.Expiry)
	case "jti":
		return reflect.ValueOf(c.Id)
	case "iat":
		return reflect.ValueOf(c.IssueAt)
	case "nbf":
		return reflect.ValueOf(c.NotBefore)
	case "iss":
		return reflect.ValueOf(c.Issuer)
	case "sub":
		return reflect.ValueOf(c.Subject)
	default:
		return reflect.Zero(reflect.TypeOf(nil))
	}
}

// customCompatibleClaims wraps customClaims and implements oauth2.Claims
type customCompatibleClaims struct {
	oauth2.FieldClaimsMapper
	customClaims
}

func (c customCompatibleClaims) Valid() error {
	return nil
}

func (c *customCompatibleClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *customCompatibleClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}