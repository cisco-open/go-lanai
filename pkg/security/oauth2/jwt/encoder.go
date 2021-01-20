package jwt

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

/*********************
	Abstract
 *********************/
type JwtEncoder interface {
	Encode(ctx context.Context, claims Claims) (string, error)
}

/*********************
	Implements
 *********************/

// RSJwtEncoder implements JwtEncoder
type RSJwtEncoder struct {
	defaultKid string
	jwkStore   JwkStore
	method     jwt.SigningMethod
}

func NewRS256JwtEncoder(jwkStore JwkStore, defaultKid string) *RSJwtEncoder {
	return &RSJwtEncoder{
		defaultKid: defaultKid,
		jwkStore: jwkStore,
		method: jwt.SigningMethodRS256,
	}
}

func (enc *RSJwtEncoder) Encode(ctx context.Context, claims Claims) (string, error) {
	token := jwt.NewWithClaims(enc.method, &jwtGoCompatibleClaims{claims: claims})
	jwk, e := enc.findJwk(ctx)
	if e != nil {
		return "", e
	}

	// set kid if not default
	if jwk.Id() != enc.defaultKid {
		token.Header[JwtHeaderKid] = jwk.Id()
	}

	return token.SignedString(jwk.Private())
}

func (enc *RSJwtEncoder) findJwk(ctx context.Context) (PrivateJwk, error) {
	switch enc.jwkStore.(type) {
	case JwkRotator:
		if jwk, e := enc.jwkStore.(JwkRotator).Current(ctx); e != nil {
			return nil, e
		} else if private, ok := jwk.(PrivateJwk); !ok {
			return nil, fmt.Errorf("current JWK doesn't have private key")
		} else {
			return private, nil
		}
	default:
		if jwk, e := enc.jwkStore.LoadByKid(ctx, enc.defaultKid); e != nil {
			return nil, e
		} else if private, ok := jwk.(PrivateJwk); !ok {
			return nil, fmt.Errorf("JWK with kid[%s] doesn't have private key", enc.defaultKid)
		} else {
			return private, nil
		}
	}
}
