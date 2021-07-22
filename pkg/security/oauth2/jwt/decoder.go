package jwt

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"github.com/dgrijalva/jwt-go"
)

/*********************
	Abstract
 *********************/

//goland:noinspection GoNameStartsWithPackageName
type JwtDecoder interface {
	Decode(ctx context.Context, token string) (oauth2.Claims, error)
	DecodeWithClaims(ctx context.Context, token string, claims interface{}) error
}

/*********************
	Implements
 *********************/

// RSJwtDecoder implements JwtEncoder
type RSJwtDecoder struct {
	jwkName  string
	jwkStore JwkStore
	parser   *jwt.Parser
}

func NewRS256JwtDecoder(jwkStore JwkStore, defaultJwkName string) *RSJwtDecoder {
	parser := &jwt.Parser{
		UseJSONNumber: false,
		SkipClaimsValidation: true,
	}
	return &RSJwtDecoder{
		jwkName:  defaultJwkName,
		jwkStore: jwkStore,
		parser:   parser,
	}
}

func (dec *RSJwtDecoder) Decode(ctx context.Context, tokenString string) (oauth2.Claims, error) {
	claims := oauth2.MapClaims{}
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
		var jwk Jwk
		var e error

		switch kid, ok := unverified.Header[JwtHeaderKid].(string); {
		case ok:
			jwk, e = dec.jwkStore.LoadByKid(ctx, kid)
		default:
			jwk, e = dec.jwkStore.LoadByName(ctx, dec.jwkName)
		}
		if e != nil {
			return nil, e
		}

		return jwk.Public(), nil
	}
}
