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
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/utils/cryptoutils"
	"hash"
)

const (
	errTmplInvalidJwkName    = `invalid JWK name`
	errTmplPubPrivMixed      = `found both public and private key block in same PEM file`
	errTmplNoKeyFoundInPem   = `PEM file doesn't includes any supported private nor public keys`
	errTmplUnsupportedPubKey = `non-supported public key [%T] in certificate`
	errTmplUnsupportedFile   = `unrecognized crypto key file format [%s]`
	errTmplUnsupportedBlock  = `non-supported block [%T] in the file`
)

// FileJwkStore implements JwkStore and JwkRotator
// This store uses load key files for public and private keys.
// File locations and "kids" are read from properties. And rotate between pre-defined keys
// The properties are structured as follows:
//
//	keys:
//	 my-key-name:
//	   id: my-key-id
//	   format: pem
//	   file: my-key-file.pem
//
// Keys loaded under the same key name will all have the same name. The LoadByName method will load one of the keys.
// Which key will be loaded is determined by the current index for that name. The Rotate method will increment the index
// for that name.
// If id property is provided, the actual key id will be the property id plus an integer suffix.
// If id property is not provided, the actual key id will be the key name plus a UUID suffix.
//
// Supports PEM format.
// Supports:
// 1. pkcs8 unencrypted private key - because golang standard library does not support this
// 2. traditional unencrypted private key and encrypted private key
// 3. traditional public key (pkcs1 rsa or pkix for ecdsa)
// 4. x509 certificate
// 5. HMAC key (using custom label "HMAC KEY", i.e. -----BEGIN HMAC KEY-----)
type FileJwkStore struct {
	cacheById   map[string]Jwk
	cacheByName map[string][]Jwk
	indexes     map[string]int
}

func NewFileJwkStore(props CryptoProperties) *FileJwkStore {
	s := FileJwkStore{
		cacheById:   map[string]Jwk{},
		cacheByName: map[string][]Jwk{},
		indexes:     map[string]int{},
	}

	// load files
	for k, v := range props.Keys {
		jwks, e := loadJwks(k, v)
		// ignore unsupported keys
		if e == nil {
			for _, jwk := range jwks {
				s.cacheById[jwk.Id()] = jwk

			}
			s.cacheByName[k] = jwks
			s.indexes[k] = 0
		} else {
			logger.Warnf("ignored key %s due to error %v", k, e)
		}
	}

	return &s
}

func (s *FileJwkStore) LoadByKid(_ context.Context, kid string) (Jwk, error) {
	jwk, ok := s.cacheById[kid]
	if !ok {
		return nil, fmt.Errorf("cannot find JWK with kid [%s]", kid)
	}
	return jwk, nil
}

func (s *FileJwkStore) LoadByName(_ context.Context, name string) (Jwk, error) {
	jwks, ok := s.cacheByName[name]
	if !ok || len(jwks) == 0 {
		return nil, fmt.Errorf("cannot find JWK with name [%s]", name)
	}

	i := s.indexes[name] % len(jwks)
	return jwks[i], nil
}

func (s *FileJwkStore) LoadAll(_ context.Context, names ...string) ([]Jwk, error) {
	jwks := make([]Jwk, 0, len(s.cacheById))

	for k, v := range s.cacheByName {
		match := len(names) == 0 // if names is empty, match all
		for i := 0; !match && i < len(names); i++ {
			match = names[i] == k
		}
		if !match {
			continue
		}

		for _, jwk := range v {
			jwks = append(jwks, jwk)
		}
	}

	return jwks, nil
}

func (s *FileJwkStore) Rotate(_ context.Context, name string) error {
	current, ok := s.indexes[name]
	if !ok {
		return fmt.Errorf(errTmplInvalidJwkName)
	}

	jwks, ok := s.cacheByName[name]
	if !ok || len(jwks) == 0 {
		return fmt.Errorf(errTmplInvalidJwkName)
	}

	s.indexes[name] = (current + 1) % len(jwks)
	return nil
}

/*************************
	Helpers
 *************************/

func loadJwks(name string, props CryptoKeyProperties) ([]Jwk, error) {
	switch props.Format() {
	case KeyFileFormatPem:
		return loadJwksFromPem(name, props)
	default:
		return nil, fmt.Errorf(errTmplUnsupportedFile, props.KeyFormat)
	}
}

func loadJwksFromPem(name string, props CryptoKeyProperties) ([]Jwk, error) {
	items, e := cryptoutils.LoadMultiBlockPem(props.Location, props.Password)
	if e != nil {
		return nil, fmt.Errorf("unable to load JWK [%s] - %v", name, e)
	}

	privJwks := make([]Jwk, 0)
	pubJwks := make([]Jwk, 0)
	for i, v := range items {
		var privKey crypto.PrivateKey
		var pubKey crypto.PublicKey

		// get private or public key
		switch v.(type) {
		case *rsa.PrivateKey:
			privKey = v.(*rsa.PrivateKey)
		case *rsa.PublicKey:
			pubKey = v.(*rsa.PublicKey)
		case *ecdsa.PrivateKey:
			privKey = v.(*ecdsa.PrivateKey)
		case *ecdsa.PublicKey:
			pubKey = v.(*ecdsa.PublicKey)
		case ed25519.PrivateKey:
			privKey = v.(ed25519.PrivateKey)
		case ed25519.PublicKey:
			pubKey = v.(ed25519.PublicKey)
		case *x509.Certificate:
			cert := v.(*x509.Certificate)
			switch cert.PublicKey.(type) {
			case *rsa.PublicKey:
				pubKey = cert.PublicKey.(*rsa.PublicKey)
			case *ecdsa.PublicKey:
				pubKey = cert.PublicKey.(*ecdsa.PublicKey)
			case ed25519.PublicKey:
				pubKey = cert.PublicKey.(ed25519.PublicKey)
			default:
				return nil, fmt.Errorf(errTmplUnsupportedPubKey, cert.PublicKey)
			}
		case *pem.Block:
			switch v.(*pem.Block).Type {
			case "HMAC KEY":
				logger.Warnf("File contains HMAC keys, please make sure the jwks end point is secured")
				privKey = v.(*pem.Block).Bytes
			default:
				return nil, fmt.Errorf(errTmplUnsupportedBlock, v)
			}
		default:
			return nil, fmt.Errorf(errTmplUnsupportedBlock, v)
		}

		// validate and create JWK
		switch {
		case privKey == nil && len(privJwks) != 0:
			return nil, fmt.Errorf(errTmplPubPrivMixed)
		case privKey == nil:
			kid := calculateKid(props, name, i, len(items), pubKey)
			pubJwks = append(pubJwks, NewJwk(kid, name, pubKey))
		case len(pubJwks) != 0:
			return nil, fmt.Errorf(errTmplPubPrivMixed)
		default:
			kid := calculateKid(props, name, i, len(items), privKey)
			privJwks = append(privJwks, NewPrivateJwk(kid, name, privKey))
		}
	}

	switch {
	case len(pubJwks) == 0 && len(privJwks) == 0:
		return nil, fmt.Errorf(errTmplNoKeyFoundInPem)
	case len(pubJwks) != 0 && len(privJwks) != 0:
		// this should not happen if previous logic (in loop) were correct
		return nil, fmt.Errorf(errTmplPubPrivMixed)
	case len(pubJwks) != 0:
		return pubJwks, nil
	case len(privJwks) != 0:
		fallthrough
	default:
		return privJwks, nil
	}
}

func calculateKid(props CryptoKeyProperties, name string, blockIndex int, numBlocks int, key any) string {
	if numBlocks == 1 {
		return name
	}

	if props.Id != "" {
		return fmt.Sprintf("%s-%d", props.Id, blockIndex)
	}

	//best effort to generate a kid that is consistent across restarts
	var hash hash.Hash
	switch key.(type) {
	case *rsa.PrivateKey:
		privKey := key.(*rsa.PrivateKey)
		hash = hashForRSA(&privKey.PublicKey)
	case *rsa.PublicKey:
		hash = hashForRSA(key.(*rsa.PublicKey))
	case *ecdsa.PrivateKey:
		privKey := key.(*ecdsa.PrivateKey)
		hash = hashForEcdsa(privKey.Public().(*ecdsa.PublicKey))
	case *ecdsa.PublicKey:
		hash = hashForEcdsa(key.(*ecdsa.PublicKey))
	case ed25519.PrivateKey:
		privKey := key.(ed25519.PrivateKey)
		hash = hashForEd25519(privKey.Public().(ed25519.PublicKey))
	case ed25519.PublicKey:
		hash = hashForEd25519(key.(ed25519.PublicKey))
	case []byte:
		hash = hmac.New(sha256.New, key.([]byte))
		hash.Write([]byte(name))
	}
	sum := hash.Sum(nil)
	suffix := hex.EncodeToString(sum)
	return name + "-" + suffix
}

func hashForRSA(key *rsa.PublicKey) hash.Hash {
	hash := sha256.New224()
	_, _ = hash.Write(key.N.Bytes())
	_ = binary.Write(hash, binary.LittleEndian, int64(key.E))
	return hash
}

func hashForEd25519(key ed25519.PublicKey) hash.Hash {
	hash := sha256.New224()
	_, _ = hash.Write(key)
	return hash
}

func hashForEcdsa(key *ecdsa.PublicKey) hash.Hash {
	hash := sha256.New224()
	_, _ = hash.Write(key.X.Bytes())
	_, _ = hash.Write(key.Y.Bytes())
	return hash
}
