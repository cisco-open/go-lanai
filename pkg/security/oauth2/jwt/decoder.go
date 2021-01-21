package jwt

import (
	"context"
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
)

/*********************
	Abstract
 *********************/
type JwtDecoder interface {
	Decode(ctx context.Context, token string) (MapClaims, error)
	DecodeWithClaims(ctx context.Context, token string, claims interface{}) error
}

/*********************
	Implements
 *********************/
// RSJwtEncoder implements JwtEncoder
type RSJwtDecoder struct {
	defaultKid string
	jwkStore   JwkStore
	parser     *jwt.Parser
}

func NewRS256JwtDecoder(jwkStore JwkStore, defaultKid string) *RSJwtDecoder {
	parser := &jwt.Parser{
		UseJSONNumber: false,
		SkipClaimsValidation: true,
	}
	return &RSJwtDecoder{
		defaultKid: defaultKid,
		jwkStore: jwkStore,
		parser: parser,
	}
}

func (dec *RSJwtDecoder) Decode(ctx context.Context, tokenString string) (MapClaims, error) {
	claims := MapClaims{}
	if e := dec.DecodeWithClaims(ctx, tokenString, &claims); e != nil {
		return nil, e
	}
	return claims, nil
}

func (dec *RSJwtDecoder) DecodeWithClaims(ctx context.Context, tokenString string, claims interface{}) (err error) {
	// type checks
	switch claims.(type) {
	case jwt.Claims:
		_, err = dec.parser.ParseWithClaims(tokenString, claims.(jwt.Claims), dec.keyFunc(ctx))
	default:
		compatible := jwtGoCompatibleClaims{
			claims: claims,
		}
		_, err = dec.parser.ParseWithClaims(tokenString, &compatible, dec.keyFunc(ctx))
	}
	return
}

func (dec *RSJwtDecoder) keyFunc(ctx context.Context) jwt.Keyfunc {
	return func(unverified *jwt.Token) (interface{}, error) {
		kid, ok := unverified.Header[JwtHeaderKid].(string)
		if !ok {
			kid = dec.defaultKid
		}

		jwk, e := dec.jwkStore.LoadByKid(ctx, kid)
		if e != nil {
			return nil, e
		}
		str := printPublicKey(jwk.Public().(*rsa.PublicKey))
		_ = ctx.Value(str)
		return jwk.Public(), nil
	}
}
