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
	Implements
 *********************/

// RSJwtDecoder implements JwtEncoder
type RSJwtDecoder struct {
	jwkName  string
	jwkStore JwkStore
	parser   *jwt.Parser
}

func NewRS256JwtDecoder(jwkStore JwkStore, defaultJwkName string) *RSJwtDecoder {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation(), jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
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
