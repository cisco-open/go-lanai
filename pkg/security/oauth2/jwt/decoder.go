package jwt

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
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
	parser := jwt.NewParser(jwt.WithoutClaimsValidation(), jwt.WithValidMethods([]string{"RS256"}))
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

// PlaintextJwtDecoder implements JwtEncoder
type PlaintextJwtDecoder struct {
	jwkName  string
	jwkStore JwkStore
	parser   *jwt.Parser
}

func NewPlaintextJwtDecoder() *PlaintextJwtDecoder {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation(), jwt.WithValidMethods([]string{"none"}))
	return &PlaintextJwtDecoder{
		parser: parser,
	}
}

func (dec *PlaintextJwtDecoder) Decode(ctx context.Context, tokenString string) (oauth2.Claims, error) {
	claims := oauth2.MapClaims{}
	if e := dec.DecodeWithClaims(ctx, tokenString, &claims); e != nil {
		return nil, e
	}
	return claims, nil
}

func (dec *PlaintextJwtDecoder) DecodeWithClaims(_ context.Context, tokenString string, claims interface{}) (err error) {
	// type checks
	switch claims.(type) {
	case jwt.Claims:
		_, err = dec.parser.ParseWithClaims(tokenString, claims.(jwt.Claims), dec.keyFunc)
	default:
		compatible := jwtGoCompatibleClaims{
			claims: claims,
		}
		_, err = dec.parser.ParseWithClaims(tokenString, &compatible, dec.keyFunc)
	}
	return
}

func (dec *PlaintextJwtDecoder) keyFunc(unverified *jwt.Token) (interface{}, error) {
	switch typ, ok := unverified.Header[JwtHeaderAlgorithm].(string); {
	case ok && typ == "none":
		return jwt.UnsafeAllowNoneSignatureType, nil
	default:
		return nil, fmt.Errorf("unsupported alg")
	}
}
