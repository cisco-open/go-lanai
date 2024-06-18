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
	"github.com/golang-jwt/jwt/v4"
)

/*********************
	Abstract
 *********************/

//goland:noinspection GoNameStartsWithPackageName
type JwtEncoder interface {
	Encode(ctx context.Context, claims interface{}) (string, error)
}

/*********************
	Constructors
 *********************/

type SigningOptions func(opt *SigningOption)
type SigningOption struct {
	JwkStore JwkStore
	JwkName  string
	Method   jwt.SigningMethod
}

// SignWithJwkStore is a SigningOptions that set JwkStore and key name to use when signing
func SignWithJwkStore(store JwkStore, jwkName string) SigningOptions {
	return func(opt *SigningOption) {
		opt.JwkStore = store
		opt.JwkName = jwkName
	}
}

// SignWithMethod is SigningOptions that specify the method to use. When set to nil, the encoder would
// attempt to use the private key type to resolve signing method.
func SignWithMethod(method jwt.SigningMethod) SigningOptions {
	return func(opt *SigningOption) {
		opt.Method = method
	}
}

// NewSignedJwtEncoder create a JwtEncoder that sign JWT with provided method.
// Depending on the sign method, provided JwkStore should supply proper private keys.
// Note: When using HS algorithms, the HMAC secret is treated as both public and private key,
//       and it would be exposed via JWKS endpoint. It is service implementer's responsibility to
//       protect the JWKS endpoint to prevent accidental leaking of HMAC secret.
func NewSignedJwtEncoder(opts ...SigningOptions) *SignedJwtEncoder {
	opt := SigningOption{
		Method: jwt.SigningMethodRS256,
	}
	for _, fn := range opts {
		fn(&opt)
	}
	return &SignedJwtEncoder{
		jwkName:  opt.JwkName,
		jwkStore: opt.JwkStore,
		method:   opt.Method,
	}
}

/*********************
	Implements
 *********************/

// SignedJwtEncoder implements JwtEncoder. It encodes claims with crypto signature of choice.
// Encoder may return error if private key is not compatible with signing method
type SignedJwtEncoder struct {
	jwkName  string
	jwkStore JwkStore
	method   jwt.SigningMethod
}

func (enc *SignedJwtEncoder) Encode(ctx context.Context, claims interface{}) (string, error) {
	// choose PrivateKey to use
	jwk, e := enc.findJwk(ctx)
	if e != nil {
		return "", e
	}

	// resolve signing method
	method := enc.method
	if method == nil {
		if method, e = resolveSigningMethod(jwk.Private()); e != nil {
			return "", e
		}
	}

	// type checks
	var token *jwt.Token
	switch claims.(type) {
	case jwt.Claims:
		token = jwt.NewWithClaims(method, claims.(jwt.Claims))
	default:
		token = jwt.NewWithClaims(method, &jwtGoCompatibleClaims{claims: claims})
	}

	// set Kid if not default
	if jwk.Id() != enc.jwkName {
		token.Header[JwtHeaderKid] = jwk.Id()
	}

	return token.SignedString(jwk.Private())
}

func (enc *SignedJwtEncoder) findJwk(ctx context.Context) (PrivateJwk, error) {
	if jwk, e := enc.jwkStore.LoadByName(ctx, enc.jwkName); e != nil {
		return nil, e
	} else if private, ok := jwk.(PrivateJwk); !ok {
		return nil, fmt.Errorf("JWK with name[%s] doesn't have private key", enc.jwkName)
	} else {
		return private, nil
	}
}
