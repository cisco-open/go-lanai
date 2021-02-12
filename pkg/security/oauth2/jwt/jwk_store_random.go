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

