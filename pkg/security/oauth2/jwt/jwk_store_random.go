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
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

// SingleJwkStore implements JwkStore
// This store always returns single JWK if kid matches, return error if not
// This store is majorly for testing
type SingleJwkStore struct {
	kid string
	jwk Jwk
}

func NewSingleJwkStore(kid string) *SingleJwkStore {
	return &SingleJwkStore{
		kid: kid,
		jwk: newJwk(kid),
	}
}

func (s *SingleJwkStore) LoadByKid(_ context.Context, kid string) (Jwk, error) {
	if s.kid == kid {
		return s.jwk, nil
	}
	return nil, fmt.Errorf("Cannot find JWK with kid [%s]", kid)
}

func (s *SingleJwkStore) LoadByName(_ context.Context, name string) (Jwk, error) {
	if s.kid == name {
		return s.jwk, nil
	}
	return nil, fmt.Errorf("Cannot find JWK with name [%s]", name)
}

func (s *SingleJwkStore) LoadAll(ctx context.Context, names ...string) ([]Jwk, error) {
	return []Jwk{s.jwk}, nil
}

func newJwk(kid string) Jwk {
	key, e := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if e != nil {
		panic(e)
	}

	return NewRsaPrivateJwk(kid, kid, key)
}

