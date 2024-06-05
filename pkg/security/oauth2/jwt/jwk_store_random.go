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
	"sync"
)

// SingleJwkStore implements JwkStore
// This store always returns single JWK if Kid matches, return error if not
// This store is majorly for testing
type SingleJwkStore struct {
	initOnce      sync.Once
	Kid           string
	SigningMethod jwt.SigningMethod
	jwk           Jwk
}

func NewSingleJwkStoreWithOptions(opts ...func(s *SingleJwkStore)) *SingleJwkStore {
	store := SingleJwkStore{
		SigningMethod: jwt.SigningMethodRS256,
	}
	for _, fn := range opts {
		fn(&store)
	}
	return &store
}

func NewSingleJwkStore(kid string) *SingleJwkStore {
	return NewSingleJwkStoreWithOptions(func(s *SingleJwkStore) {
		s.Kid = kid
	})
}

func (s *SingleJwkStore) LoadByKid(_ context.Context, kid string) (Jwk, error) {
	if e := s.LazyInit(); e != nil {
		return nil, e
	}
	if s.Kid == kid {
		return s.jwk, nil
	}
	return nil, fmt.Errorf("cannot find JWK with Kid [%s]", kid)
}

func (s *SingleJwkStore) LoadByName(_ context.Context, name string) (Jwk, error) {
	if e := s.LazyInit(); e != nil {
		return nil, e
	}
	if s.Kid == name {
		return s.jwk, nil
	}
	return nil, fmt.Errorf("cannot find JWK with name [%s]", name)
}

func (s *SingleJwkStore) LoadAll(_ context.Context, _ ...string) ([]Jwk, error) {
	if e := s.LazyInit(); e != nil {
		return nil, e
	}
	return []Jwk{s.jwk}, nil
}

func (s *SingleJwkStore) LazyInit() (err error) {
	s.initOnce.Do(func() {
		s.jwk, err = generateRandomJwk(s.SigningMethod, s.Kid, s.Kid)
		if err != nil {
			return
		}
	})
	return
}
