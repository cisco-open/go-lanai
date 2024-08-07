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
	"crypto"
	"encoding/json"
	"reflect"
)

/*********************
	Abstraction
 *********************/

type Jwk interface {
	Id() string
	Name() string
	Public() crypto.PublicKey
}

type PrivateJwk interface {
	Jwk
	Private() crypto.PrivateKey
}

type JwkStore interface {
	// LoadByKid returns the JWK associated with given KID.
	// This method is usually used when decoding/verifiying JWT token
	LoadByKid(ctx context.Context, kid string) (Jwk, error)

	// LoadByName returns the JWK associated with given name.
	// The method might return different JWK for same name, if the store is also support rotation
	// This method is usually used when encoding/encrypt JWT token
	// Note: if the store does not support rotation (i.e. it doest not implement JwkRotator),
	// this store could use the name as the jwk id. Doing so would allow the encoder to not
	// add a "kid" header to the JWT token. This allows the use case where the JWT key is agreed upon by
	// both the encoder and decoder through an out-of-band mechanism without using "kid".
	// See the comment in SignedJwtEncoder.Encode for more details
	LoadByName(ctx context.Context, name string) (Jwk, error)

	// LoadAll return all JWK with given names. If name is not provided, all JWK is returned
	LoadAll(ctx context.Context, names ...string) ([]Jwk, error)
}

type JwkRotator interface {
	JwkStore
	// Rotate change JWK of given name to next candicate
	Rotate(ctx context.Context, name string) error
}

/*********************
	Implements Base
 *********************/

// GenericJwk implements Jwk
type GenericJwk struct {
	kid    string
	name   string
	public crypto.PublicKey
}

func (k *GenericJwk) Id() string {
	return k.kid
}

func (k *GenericJwk) Name() string {
	return k.name
}

func (k *GenericJwk) Public() crypto.PublicKey {
	return k.public
}

func (k *GenericJwk) MarshalJSON() ([]byte, error) {
	return marshalJwk(k)
}

func (k *GenericJwk) UnmarshalJSON(data []byte) error {
	jwk, e := unmarshalJwk(data)
	if e != nil {
		return e
	}
	switch v := jwk.(type) {
	case *GenericJwk:
		*k = *v
	default:
		*k = GenericJwk{kid: jwk.Id(), name: jwk.Name(), public: jwk.Public()}
	}
	return nil
}

// GenericPrivateJwk implements Jwk and PrivateJwk
type GenericPrivateJwk struct {
	GenericJwk
	private crypto.PrivateKey
}

func (k *GenericPrivateJwk) Private() crypto.PrivateKey {
	return k.private
}

/*********************
	Constructors
 *********************/

var typeOfBytes = reflect.TypeOf((*[]byte)(nil)).Elem()

type publicKey interface {
	Equal(x crypto.PublicKey) bool
}

type privateKey interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
}

// NewJwk new Jwk with specified public key
// Supported public key types:
//   - *rsa.PublicKey
//   - *ecdsa.PublicKey
//   - ed25519.PublicKey
//   - []byte (MAC secret)
//   - any key implementing:
//     interface{
//     Equal(x crypto.PublicKey) bool
//     }
func NewJwk(kid string, name string, pubKey crypto.PublicKey) Jwk {
	return &GenericJwk{
		kid:    kid,
		name:   name,
		public: pubKey,
	}
}

// ParseJwk parse Jwk from JSON as specified in RFC 7517 and RFC 7518.
// Note: Private key information is ignored in the parsed Jwk.
// Supported public key types:
// - *rsa.PublicKey (kty = RSA)
// - *ecdsa.PublicKey (kty = EC)
// - ed25519.PublicKey (kty = OKP)
// - []byte (symmetric key, e.g. MAC secret) ((kty = oct)
//
// See: RFC7517 https://datatracker.ietf.org/doc/html/rfc7517
// See: RFC7518 https://datatracker.ietf.org/doc/html/rfc7518
func ParseJwk(jsonData []byte) (Jwk, error) {
	var jwk GenericJwk
	if e := json.Unmarshal(jsonData, &jwk); e != nil {
		return nil, e
	}
	return &jwk, nil
}

// NewPrivateJwk new PrivateJwk with specified private key
// Supported private key types:
//   - *rsa.PrivateKey
//   - *ecdsa.PrivateKey
//   - ed25519.PrivateKey
//   - []byte (MAC secret)
//   - any key implementing:
//     interface{
//     Public() crypto.PublicKey
//     Equal(x crypto.PrivateKey) bool
//     }
func NewPrivateJwk(kid string, name string, privKey crypto.PrivateKey) PrivateJwk {
	var pubKey crypto.PublicKey
	switch v := privKey.(type) {
	case privateKey:
		pubKey = v.Public()
	default:
		// HMAC secret
		if rv := reflect.ValueOf(privKey); rv.CanConvert(typeOfBytes) {
			privKey = rv.Convert(typeOfBytes).Interface()
			pubKey = privKey
		}
	}
	return &GenericPrivateJwk{
		GenericJwk: GenericJwk{
			kid:    kid,
			name:   name,
			public: pubKey,
		},
		private: privKey,
	}
}
