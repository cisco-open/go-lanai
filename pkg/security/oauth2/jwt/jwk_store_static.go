package jwt

import (
	"context"
	"crypto/rsa"
	"hash/crc64"
	"math/rand"
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

func (s *StaticJwkStore) Rotate(ctx context.Context) error {
	s.current = (s.current + 1) % len(kidRoundRobin)
	return nil
}

func (s *StaticJwkStore) Current(ctx context.Context) (Jwk, error) {
	return s.LoadByKid(ctx, kidRoundRobin[s.current])
}

func (s *StaticJwkStore) LoadByKid(_ context.Context, kid string) (Jwk, error) {
	return s.getOrNew(kid)
}

func (s *StaticJwkStore) LoadByName(_ context.Context, name string) (Jwk, error) {
	return s.getOrNew(name)
}

func (s *StaticJwkStore) LoadAll(_ context.Context) ([]Jwk, error) {
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

	reader := newPseudoReader(kid)
	key, e := rsa.GenerateKey(reader, rsaKeySize)
	if e != nil {
		return nil, e
	}

	jwk := NewRsaPrivateJwk(kid, kid, key)
	s.lookup[kid] = jwk

	return jwk, nil
}

// pseudoReader implements io.Reader
type pseudoReader struct {
	rand *rand.Rand
}

func newPseudoReader(seed string) pseudoReader {
	table := crc64.MakeTable(crc64.ISO)
	crc := crc64.Checksum([]byte(seed), table)
	source := rand.NewSource(int64(crc))
	return pseudoReader{
		rand: rand.New(source),
	}

}

func (r pseudoReader) Read(p []byte) (n int, err error) {
	return r.rand.Read(p)
}
