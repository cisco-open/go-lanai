package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	DefaultKidSuffix = "-default"
)

// FileJwkStore implements JwkStore and JwkRotator
// This store uses load key files for public and private keys.
// File locations and "kids" are read from properties. And rotate between pre-defined keys
type FileJwkStore struct {
	cacheById   map[string]Jwk
	cacheByName map[string]Jwk
	kids        []string
	current     int
}

func NewFileJwkStore(props CryptoProperties) *FileJwkStore {
	s := FileJwkStore{
		cacheById: map[string]Jwk{},
		cacheByName: map[string]Jwk{},
		kids: []string{},
	}

	// load files
	for k, v := range props.Keys {
		jwk, e := loadJwk(k, v)
		if e != nil {
			panic(e)
		}
		s.cacheById[jwk.Id()] = jwk
		s.cacheByName[jwk.Name()] = jwk
		s.kids = append(s.kids, jwk.Id())
	}

	return &s
}

func (s *FileJwkStore) LoadByKid(ctx context.Context, kid string) (Jwk, error) {
	jwk, ok := s.cacheById[kid]
	if !ok {
		return nil, fmt.Errorf("cannot find JWK with kid [%s]", kid)
	}
	return jwk, nil
}

func (s *FileJwkStore) LoadByName(ctx context.Context, name string) (Jwk, error) {
	jwk, ok := s.cacheByName[name]
	if !ok {
		return nil, fmt.Errorf("cannot find JWK with name [%s]", name)
	}
	return jwk, nil
}

func (s *FileJwkStore) LoadAll(ctx context.Context) ([]Jwk, error) {
	jwks := make([]Jwk, len(s.cacheById))

	i := 0
	for _, v := range s.cacheById {
		jwks[i] = v
		i++
	}
	return jwks, nil
}

func (s *FileJwkStore) Rotate(ctx context.Context) error {
	s.current = (s.current + 1) % len(s.kids)
	return nil
}

func (s *FileJwkStore) Current(ctx context.Context) (Jwk, error) {
	return s.LoadByKid(ctx, s.kids[s.current])
}

/*************************
	Helpers
 *************************/
func loadJwk(name string, props CryptoKeyProperties) (Jwk, error) {
	switch props.Format() {
	case KeyFileFormatPem:
		return loadJwkFromPem(name, props)
	default:
		return nil, fmt.Errorf("Unrecognized crypto key file format [%s]", props.KeyFormat)
	}
	return nil, nil
}

func loadJwkFromPem(name string, props CryptoKeyProperties) (Jwk, error) {
	items, e := cryptoutils.LoadMultiBlockPem(props.Location, props.Password)
	if e != nil {
		return nil, fmt.Errorf("unable to load JWK [%s] - %v", name, e)
	}

	var privKey *rsa.PrivateKey
	var pubKey *rsa.PublicKey
	for _, v := range items {
		switch v.(type) {
		case *rsa.PrivateKey:
			if privKey != nil {
				return nil, fmt.Errorf("found multiple private keys in the file")
			}
			privKey = v.(*rsa.PrivateKey)
		case *rsa.PublicKey:
			if pubKey != nil {
				return nil, fmt.Errorf("found multiple public keys/certs in the file")
			}
			pubKey = v.(*rsa.PublicKey)
		case *x509.Certificate:
			cert := v.(*x509.Certificate)
			k, ok := cert.PublicKey.(*rsa.PublicKey)
			if !ok {
				return nil, fmt.Errorf("non-supported public key [%T] in certificate", cert.PublicKey)
			} else if pubKey != nil {
				return nil, fmt.Errorf("found multiple public keys/certs in the file")
			}
			pubKey = k
		default:
			return nil, fmt.Errorf("non-supported block [%T] in the file", v)
		}
	}

	if pubKey == nil && privKey == nil {
		return nil, fmt.Errorf("public key is missing")
	}

	if privKey == nil {
		kid := calculateKid(props, name, pubKey)
		return NewRsaJwk(kid, name, pubKey), nil
	} else if pubKey != nil && !privKey.PublicKey.Equal(pubKey) {
		return nil, fmt.Errorf("found both public and private key block, but they don't match")
	}

	kid := calculateKid(props, name, &privKey.PublicKey)
	return NewRsaPrivateJwk(kid, name, privKey), nil
}

func calculateKid(props CryptoKeyProperties, name string, key *rsa.PublicKey) string {
	if props.Id != "" {
		return props.Id
	}

	// best effort to create a unique suffix for the kid
	hash := sha256.New224()
	hash.Write(key.N.Bytes())
	binary.Write(hash, binary.LittleEndian, key.E)
	sum := hash.Sum(nil)
	suffix := hex.EncodeToString(sum)

	return name + "-" + suffix
}