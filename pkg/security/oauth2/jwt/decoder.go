package jwt

import (
	"context"
	"github.com/dgrijalva/jwt-go"
)

/*********************
	Abstract
 *********************/
type JwtDecoder interface {
	Decode(ctx context.Context, token string, claims Claims) (Claims, error)
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

func (dec *RSJwtDecoder) Decode(ctx context.Context, tokenString string, claims Claims) (Claims, error) {
	compatible := jwtGoCompatibleClaims{
		claims: claims,
	}

	_, e := jwt.ParseWithClaims(tokenString, &compatible, dec.keyFunc(ctx))
	if e != nil {
		return nil, e
	}
	return compatible.claims, nil
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
		return jwk.Public(), nil
	}
}
