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
	Encode(ctx context.Context, claims interface{}) (string, error)
}

/*********************
	Implements
 *********************/

// RSJwtEncoder implements JwtEncoder
type RSJwtEncoder struct {
	jwkName  string
	jwkStore JwkStore
	method   jwt.SigningMethod
}

func NewRS256JwtEncoder(jwkStore JwkStore, jwkName string) *RSJwtEncoder {
	return &RSJwtEncoder{
		jwkName:  jwkName,
		jwkStore: jwkStore,
		method:   jwt.SigningMethodRS256,
	}
}

func (enc *RSJwtEncoder) Encode(ctx context.Context, claims interface{}) (string, error) {
	// type checks
	var token *jwt.Token
	switch claims.(type) {
	case jwt.Claims:
		token = jwt.NewWithClaims(enc.method, claims.(jwt.Claims))
	default:
		token = jwt.NewWithClaims(enc.method, &jwtGoCompatibleClaims{claims: claims})
	}

	// choose PrivateKey to use
	jwk, e := enc.findJwk(ctx)
	if e != nil {
		return "", e
	}

	// set kid if not default
	if jwk.Id() != enc.jwkName {
		token.Header[JwtHeaderKid] = jwk.Id()
	}

	return token.SignedString(jwk.Private())
}

func (enc *RSJwtEncoder) findJwk(ctx context.Context) (PrivateJwk, error) {
	if jwk, e := enc.jwkStore.LoadByName(ctx, enc.jwkName); e != nil {
		return nil, e
	} else if private, ok := jwk.(PrivateJwk); !ok {
		return nil, fmt.Errorf("JWK with name[%s] doesn't have private key", enc.jwkName)
	} else {
		return private, nil
	}
}
