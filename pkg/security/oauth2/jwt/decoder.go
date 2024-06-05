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
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
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
	Constructors
 *********************/

var allSigningMethods = []jwt.SigningMethod{
	jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512,
	jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512,
	jwt.SigningMethodES256, jwt.SigningMethodES384, jwt.SigningMethodES512,
	jwt.SigningMethodPS256, jwt.SigningMethodPS384, jwt.SigningMethodPS512,
	jwt.SigningMethodEdDSA,
}

type VerifyOptions func(opt *VerifyOption)
type VerifyOption struct {
	JwkStore      JwkStore
	JwkName       string
	Methods       []jwt.SigningMethod
	ParserOptions []jwt.ParserOption
}

// VerifyWithJwkStore is a VerifyOptions that set JwkStore and default key name to use when verifying.
// the provided key name is used as fallback if the to-be-verified JWT doesn't have "kid" in header
func VerifyWithJwkStore(store JwkStore, jwkName string) VerifyOptions {
	return func(opt *VerifyOption) {
		opt.JwkStore = store
		opt.JwkName = jwkName
	}
}

// VerifyWithMethods is a VerifyOptions that specify all allowed signing method ("alg" header).
// By default, it accepts all available signing methods except for plaintext JWT.
func VerifyWithMethods(methods ...jwt.SigningMethod) VerifyOptions {
	return func(opt *VerifyOption) {
		opt.Methods = methods
	}
}

func NewSignedJwtDecoder(opts ...VerifyOptions) *SignedJwtDecoder {
	opt := VerifyOption{
		Methods:       allSigningMethods,
		ParserOptions: []jwt.ParserOption{jwt.WithoutClaimsValidation()},
	}
	for _, fn := range opts {
		fn(&opt)
	}
	methods := make([]string, len(opt.Methods))
	for i := range opt.Methods {
		methods[i] = opt.Methods[i].Alg()
	}
	parserOpts := append(opt.ParserOptions, jwt.WithValidMethods(methods))
	return &SignedJwtDecoder{
		jwkName:  opt.JwkName,
		jwkStore: opt.JwkStore,
		parser:   jwt.NewParser(parserOpts...),
	}
}

/*********************
	Implements
 *********************/

// SignedJwtDecoder implements JwtEncoder
type SignedJwtDecoder struct {
	jwkName  string
	jwkStore JwkStore
	parser   *jwt.Parser
}

func (dec *SignedJwtDecoder) Decode(ctx context.Context, tokenString string) (oauth2.Claims, error) {
	claims := oauth2.MapClaims{}
	if e := dec.DecodeWithClaims(ctx, tokenString, &claims); e != nil {
		return nil, e
	}
	return claims, nil
}

func (dec *SignedJwtDecoder) DecodeWithClaims(ctx context.Context, tokenString string, claims interface{}) (err error) {
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

func (dec *SignedJwtDecoder) keyFunc(ctx context.Context) jwt.Keyfunc {
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
	parser := jwt.NewParser(jwt.WithoutClaimsValidation(), jwt.WithValidMethods([]string{jwt.SigningMethodNone.Alg()}))
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
