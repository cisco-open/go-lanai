// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"context"
	"errors"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/test"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"reflect"
	"testing"
	"time"
)

const (
	testDefaultKid = "default"
)

var claims = oauth2.MapClaims{
	"aud": []interface{}{"target"},
	"exp": time.Now().Add(24 * time.Hour).Unix(),
	"jti": uuid.New().String(),
	"iat": time.Now().Unix(),
	"nbf": time.Now().Unix(),
	"iss": "sandbox",
	"sub": "user",
}

/*************************
	Test Cases
 *************************/

func TestJwtWithKid(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodHS256), "HS256"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodHS384), "HS384"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodHS512), "HS512"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodRS256), "RS256"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodRS384), "RS384"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodRS512), "RS512"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodES256), "ES256"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodES384), "ES384"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodES512), "ES512"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodPS256), "PS256"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodPS384), "PS384"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodPS512), "PS512"),
		test.SubTest(SubTestJwtWithKid(jwt.SigningMethodEdDSA), "EdDSA"),
	)
}

func TestJwtWithoutKid(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestJwtWithoutKid(jwt.SigningMethodHS256), "HS256"),
		test.GomegaSubTest(SubTestJwtWithoutKid(jwt.SigningMethodRS256), "RS256"),
		test.GomegaSubTest(SubTestJwtWithoutKid(jwt.SigningMethodES256), "ES256"),
	)
}

func TestPlainJwt(t *testing.T) {
	nonRotatingJwkStore := NewSingleJwkStoreWithOptions(func(s *SingleJwkStore) {
		s.Kid = testDefaultKid
	})
	enc := NewSignedJwtEncoder(SignWithJwkStore(nonRotatingJwkStore, testDefaultKid), SignWithMethod(jwt.SigningMethodRS256))

	// encoding
	value, err := enc.Encode(context.Background(), claims)
	g := NewWithT(t)
	g.Expect(err).NotTo(HaveOccurred(), "Encode shouldn't returns error")
	g.Expect(value).NotTo(BeZero(), "Encoded jwt shouldn't be empty")

	// plain text encoding
	plainEnc := newPlainJwtEncoder()
	plainValue, err := plainEnc.Encode(context.Background(), claims)
	g.Expect(err).NotTo(HaveOccurred(), "Encode shouldn't return error")
	g.Expect(plainValue).NotTo(BeZero(), "Encoded jwt shouldn't be empty")

	t.Logf("JWT: %s", plainValue)

	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestPlainJwdDecodeSucceeds(plainValue), "SubTestPlainJwdDecodeSucceeds"),
		test.GomegaSubTest(SubTestPlainJwtDecodeFailsWithWrongAlg(value), "SubTestPlainJwtDecodeFailsWithWrongAlg"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestSetupEncodeJwt(method jwt.SigningMethod, jwkStore JwkStore, kid string) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		g := gomega.NewWithT(t)
		enc := NewSignedJwtEncoder(SignWithJwkStore(jwkStore, kid), SignWithMethod(method))

		// encoding
		value, err := enc.Encode(ctx, claims)
		g.Expect(err).NotTo(HaveOccurred(), "Encode shouldn't returns error")
		g.Expect(value).NotTo(BeZero(), "Encoded jwt shouldn't be empty")
		t.Logf("JWT: %s", value)
		headers, e := ParseJwtHeaders(value)
		g.Expect(e).To(Succeed(), "parsing JWT header should not fail")
		g.Expect(headers).To(HaveKeyWithValue("alg", method.Alg()), "encoded JWT should have correct 'alg' in header")
		return contextWithJwt(ctx, value), nil
	}
}

func SubTestJwtWithKid(method jwt.SigningMethod) test.SubTestFunc {
	return func(ctx context.Context, t *testing.T) {
		kids := []string{"kid1", "kid2", "kid3"}
		staticJwkStore := NewStaticJwkStoreWithOptions(func(s *StaticJwkStore) {
			s.KIDs = kids
			s.SigningMethod = method
		})

		test.RunTest(ctx, t,
			test.SubTestSetup(SubTestSetupEncodeJwt(method, staticJwkStore, testDefaultKid)),
			// decode, happy path
			test.GomegaSubTest(SubTestJwtEncodeWithDynamicMethod(staticJwkStore, testDefaultKid), "JwtEncodeWithDynamicMethod"),
			test.GomegaSubTest(SubTestJwtDecodeSuccessWithSameKey(staticJwkStore), "JwtDecodeSuccessWithSameKey"),
			test.GomegaSubTest(SubTestJwtDecodeSuccessWithRotatedKey(staticJwkStore), "JwtDecodeSuccessWithRotatedKey"),
			test.GomegaSubTest(SubTestJwtDecodeSuccessWithCustomClaims(staticJwkStore), "JwtDecodeSuccessWithCustomClaims"),
			// decode, not so happy, Kid exists, but not same key
			test.GomegaSubTest(SubTestJwtDecodeFailedWithWrongKey(method, kids[0]), "JwtDecodeFailedWithWrongKey"),
			test.GomegaSubTest(SubTestJwtDecodeFailedWithNonExistingKey(), "JwtDecodeFailedWithNonExistingKey"),
			// decode, not happy, alg is not supported by the decoder
			test.GomegaSubTest(SubTestJwtDecodeFailsWithWrongAlg(staticJwkStore), "JwtDecodeFailsWithWrongAlg"),
		)
	}
}

func SubTestJwtWithoutKid(method jwt.SigningMethod) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *WithT) {
		// Note, when using default "kid" defined in Encoder, "kid" field is omitted in the JWT
		nonRotatingJwkStore := NewSingleJwkStoreWithOptions(func(s *SingleJwkStore) {
			s.Kid = testDefaultKid
			s.SigningMethod = method
		})

		test.RunTest(ctx, t,
			test.SubTestSetup(SubTestSetupEncodeJwt(method, nonRotatingJwkStore, testDefaultKid)),
			// decode, happy path
			test.GomegaSubTest(SubTestJwtEncodeWithDynamicMethod(nonRotatingJwkStore, testDefaultKid), "JwtEncodeWithDynamicMethod"),
			test.GomegaSubTest(SubTestJwtDecodeSuccessWithSameKey(nonRotatingJwkStore), "JwtDecodeSuccessWithSameKey"),
			test.GomegaSubTest(SubTestJwtDecodeSuccessWithCustomClaims(nonRotatingJwkStore), "nonRotatingJwkStore"),
			// decode, not so happy, Kid exists, but not same key
			test.GomegaSubTest(SubTestJwtDecodeFailedWithWrongKey(method, testDefaultKid), "JwtDecodeFailedWithWrongKey"),
			test.GomegaSubTest(SubTestJwtDecodeFailedWithNonExistingKey(), "JwtDecodeFailedWithNonExistingKey"),
			// decode, not happy, alg is not supported by the decoder
			test.GomegaSubTest(SubTestJwtDecodeFailsWithWrongAlg(nonRotatingJwkStore), "JwtDecodeFailsWithWrongAlg"),
		)
	}
}

func SubTestJwtEncodeWithDynamicMethod(jwkStore JwkStore, kid string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		enc := NewSignedJwtEncoder(SignWithJwkStore(jwkStore, kid), SignWithMethod(nil))

		// encoding
		value, err := enc.Encode(ctx, claims)
		g.Expect(err).NotTo(HaveOccurred(), "Encode shouldn't returns error")
		g.Expect(value).NotTo(BeZero(), "Encoded jwt shouldn't be empty")
	}
}

func SubTestJwtDecodeSuccessWithSameKey(jwkStore JwkStore) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		dec := NewSignedJwtDecoder(VerifyWithJwkStore(jwkStore, testDefaultKid), VerifyWithMethods(SupportedSigningMethods...))
		parsed, err := dec.Decode(context.Background(), jwtFromContext(ctx))

		assertDecodeResult(g, parsed, err)
		assertMapClaims(g, claims, parsed)
	}
}

func SubTestJwtDecodeSuccessWithRotatedKey(jwkStore JwkRotator) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		if err := jwkStore.Rotate(context.Background(), ""); err != nil {
			t.Errorf("StaticJwkStore key roation should not have error, but got %v", err)
		}
		dec := NewSignedJwtDecoder(VerifyWithJwkStore(jwkStore, testDefaultKid), VerifyWithMethods(SupportedSigningMethods...))
		parsed, err := dec.Decode(context.Background(), jwtFromContext(ctx))

		assertDecodeResult(g, parsed, err)
		assertMapClaims(g, claims, parsed)
	}
}

func SubTestJwtDecodeSuccessWithCustomClaims(jwkStore JwkStore) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		dec := NewSignedJwtDecoder(VerifyWithJwkStore(jwkStore, testDefaultKid), VerifyWithMethods(SupportedSigningMethods...))

		custom := customClaims{}
		err := dec.DecodeWithClaims(context.Background(), jwtFromContext(ctx), &custom)
		assertDecodeResult(g, custom, err)
		assertCustomClaims(g, claims, custom)

		compatible := customCompatibleClaims{}
		err = dec.DecodeWithClaims(context.Background(), jwtFromContext(ctx), &compatible)
		assertDecodeResult(g, custom, err)
		assertCustomClaims(g, claims, compatible.customClaims)
	}
}

func SubTestJwtDecodeFailedWithWrongKey(method jwt.SigningMethod, kid string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		store := NewSingleJwkStoreWithOptions(func(s *SingleJwkStore) {
			s.Kid = kid
			s.SigningMethod = method
		})
		dec := NewSignedJwtDecoder(VerifyWithJwkStore(store, testDefaultKid), VerifyWithMethods(SupportedSigningMethods...))
		_, err := dec.Decode(context.Background(), jwtFromContext(ctx))
		g.Expect(err).NotTo(Succeed(), "decode with different JWK should return validation error")
	}
}

func SubTestJwtDecodeFailedWithNonExistingKey() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		store := NewSingleJwkStoreWithOptions(func(s *SingleJwkStore) {
			s.Kid = "whatever"
		})
		dec := NewSignedJwtDecoder(VerifyWithJwkStore(store, testDefaultKid), VerifyWithMethods(SupportedSigningMethods...))
		_, err := dec.Decode(context.Background(), jwtFromContext(ctx))
		g.Expect(err).NotTo(Succeed(), "decode with non-existing JWK should return validation error")
	}
}

func SubTestJwtDecodeFailsWithWrongAlg(jwkStore JwkStore) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {

		// plain text encoding
		plainEnc := newPlainJwtEncoder()
		plainValue, err := plainEnc.Encode(context.Background(), claims)
		g.Expect(err).NotTo(HaveOccurred(), "Encode shouldn't return error")
		g.Expect(plainValue).NotTo(BeZero(), "Encoded jwt shouldn't be empty")

		// another decorder
		dec := NewSignedJwtDecoder(VerifyWithJwkStore(jwkStore, testDefaultKid), VerifyWithMethods(SupportedSigningMethods...))
		_, e := dec.Decode(context.Background(), plainValue)

		var validationError *jwt.ValidationError
		g.Expect(errors.As(e, &validationError)).To(BeTrue())
		g.Expect(validationError.Is(jwt.ErrTokenSignatureInvalid))
	}
}

func SubTestPlainJwdDecodeSucceeds(jwtVal string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		dec := NewPlaintextJwtDecoder()
		parsed, err := dec.Decode(context.Background(), jwtVal)

		assertDecodeResult(g, parsed, err)
		assertMapClaims(g, claims, parsed)
	}
}

func SubTestPlainJwtDecodeFailsWithWrongAlg(jwtVal string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		dec := NewPlaintextJwtDecoder()
		_, err := dec.Decode(context.Background(), jwtVal)

		var validationError *jwt.ValidationError
		g.Expect(errors.As(err, &validationError)).To(BeTrue())
		g.Expect(validationError.Is(jwt.ErrTokenSignatureInvalid))
	}
}

/*************************
	Helpers
 *************************/

type ckJwt struct{}

func contextWithJwt(ctx context.Context, jwt string) context.Context {
	return context.WithValue(ctx, ckJwt{}, jwt)
}

func jwtFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ckJwt{}).(string)
	return v
}

func assertDecodeResult(g *WithT, decoded oauth2.Claims, err error) {
	g.Expect(err).NotTo(HaveOccurred(), "Decode should not return error.")
	g.Expect(decoded).NotTo(BeNil(), "Decode should return non-nil claims")
}

func assertMapClaims(g *WithT, expected oauth2.MapClaims, decoded oauth2.Claims) {

	g.Expect(decoded).To(BeAssignableToTypeOf(oauth2.MapClaims{}), "MapClaims is expected")
	actual := decoded.(oauth2.MapClaims)

	g.Expect(len(actual)).To(Equal(len(expected)), "actual MapClaims should have same size")
	for k, v := range actual {
		g.Expect(v).To(BeEquivalentTo(expected[k]), "actual MapClaims should have same [%s]", k)
	}
}

func assertCustomClaims(g *WithT, expected oauth2.MapClaims, decoded oauth2.Claims) {

	g.Expect(decoded).To(BeAssignableToTypeOf(customClaims{}), "custom claims is expected")
	actual := decoded.(customClaims)

	for k, v := range expected {
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

type plainJwtEncoder struct {
}

func newPlainJwtEncoder() *plainJwtEncoder {
	return &plainJwtEncoder{}
}

func (p *plainJwtEncoder) Encode(_ context.Context, claims interface{}) (string, error) {
	// type checks
	var token *jwt.Token
	switch claims.(type) {
	case jwt.Claims:
		token = jwt.NewWithClaims(jwt.SigningMethodNone, claims.(jwt.Claims))
	default:
		token = jwt.NewWithClaims(jwt.SigningMethodNone, &jwtGoCompatibleClaims{claims: claims})
	}

	return token.SignedString(jwt.UnsafeAllowNoneSignatureType)
}
