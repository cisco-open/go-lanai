package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
)

var (
	kidRoundRobin = []string{"kid1", "kid2", "kid3"}
)

// StaticJwkStore implements JwkStore and JwkRotator
// This store uses "kid" as seed to generate PrivateJwk. For same "kid" the returned key is same
// this one is not thread safe
type StaticJwkStore struct {
	kids    []string
	current int
	lookup  map[string]Jwk
}

func NewStaticJwkStore(kids...string) *StaticJwkStore {
	if len(kids) == 0 {
		kids = kidRoundRobin
	}
	return &StaticJwkStore{
		kids: kids,
		lookup: map[string]Jwk{},
	}
}

func (s *StaticJwkStore) Rotate(ctx context.Context, name string) error {
	s.current = (s.current + 1) % len(kidRoundRobin)
	return nil
}

func (s *StaticJwkStore) LoadByKid(_ context.Context, kid string) (Jwk, error) {
	return s.getOrNew(kid)
}

func (s *StaticJwkStore) LoadByName(_ context.Context, name string) (Jwk, error) {
	return s.getOrNew(kidRoundRobin[s.current])
}

func (s *StaticJwkStore) LoadAll(ctx context.Context, names ...string) ([]Jwk, error) {
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

	key, e := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if e != nil {
		return nil, e
	}

	jwk := NewRsaPrivateJwk(kid, kid, key)
	s.lookup[kid] = jwk

	return jwk, nil
}
