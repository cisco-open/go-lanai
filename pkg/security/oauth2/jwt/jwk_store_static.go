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
	"github.com/golang-jwt/jwt/v4"
)

var (
	kidRoundRobin = []string{"kid1", "kid2", "kid3"}
)

// StaticJwkStore implements JwkStore and JwkRotator
// This store uses "kid" as seed to generate PrivateJwk. For same "kid" the returned key is same.
// this one is not thread safe
type StaticJwkStore struct {
	KIDs          []string
	SigningMethod jwt.SigningMethod
	current       int
	lookup        map[string]Jwk
}

func NewStaticJwkStoreWithOptions(opts ...func(s *StaticJwkStore)) *StaticJwkStore {
	store := StaticJwkStore{
		KIDs:          kidRoundRobin,
		SigningMethod: jwt.SigningMethodRS256,
		lookup:        map[string]Jwk{},
	}
	for _, fn := range opts {
		fn(&store)
	}
	return &store
}

func NewStaticJwkStore(kids ...string) *StaticJwkStore {
	return NewStaticJwkStoreWithOptions(func(s *StaticJwkStore) {
		if len(kids) != 0 {
			s.KIDs = kids
		}
	})
}

func (s *StaticJwkStore) Rotate(_ context.Context, _ string) error {
	s.current = (s.current + 1) % len(s.KIDs)
	return nil
}

func (s *StaticJwkStore) LoadByKid(_ context.Context, kid string) (Jwk, error) {
	return s.getOrNew(kid)
}

func (s *StaticJwkStore) LoadByName(_ context.Context, _ string) (Jwk, error) {
	return s.getOrNew(s.KIDs[s.current])
}

func (s *StaticJwkStore) LoadAll(_ context.Context, _ ...string) ([]Jwk, error) {
	jwks := make([]Jwk, len(s.lookup))

	i := 0
	for _, v := range s.lookup {
		jwks[i] = v
		i++
	}
	return jwks, nil
}

func (s *StaticJwkStore) getOrNew(kid string) (Jwk, error) {
	if jwk, ok := s.lookup[kid]; ok {
		return jwk, nil
	}

	jwk, e := generateRandomJwk(s.SigningMethod, kid, kid)
	if e != nil {
		return nil, e
	}
	s.lookup[kid] = jwk
	return jwk, nil
}
