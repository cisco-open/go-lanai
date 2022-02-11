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
			match = match || names[i] == k
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
		var privKey *rsa.PrivateKey
		var pubKey *rsa.PublicKey

		// get private or public key
		switch v.(type) {
		case *rsa.PrivateKey:
			privKey = v.(*rsa.PrivateKey)
		case *rsa.PublicKey:
			pubKey = v.(*rsa.PublicKey)
		case *x509.Certificate:
			cert := v.(*x509.Certificate)
			k, ok := cert.PublicKey.(*rsa.PublicKey)
			if !ok {
				return nil, fmt.Errorf(errTmplUnsupportedPubKey, cert.PublicKey)
			}
			pubKey = k
		default:
			return nil, fmt.Errorf(errTmplUnsupportedBlock, v)
		}

		// validate and create JWK
		switch {
		case privKey == nil && len(privJwks) != 0:
			return nil, fmt.Errorf(errTmplPubPrivMixed)
		case privKey == nil:
			kid := calculateKid(props, name, i, pubKey)
			pubJwks = append(pubJwks, NewRsaJwk(kid, name, pubKey))
		case len(pubJwks) != 0:
			return nil, fmt.Errorf(errTmplPubPrivMixed)
		default:
			kid := calculateKid(props, name, i, &privKey.PublicKey)
			privJwks = append(privJwks, NewRsaPrivateJwk(kid, name, privKey))
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

func calculateKid(props CryptoKeyProperties, name string, blockIndex int, key *rsa.PublicKey) string {
	if props.Id != "" {
		return fmt.Sprintf("%s-%d", props.Id, blockIndex)
	}

	// best effort to create a unique suffix for the kid
	hash := sha256.New224()
	_, _ = hash.Write(key.N.Bytes())
	_ = binary.Write(hash, binary.LittleEndian, key.E)
	sum := hash.Sum(nil)
	suffix := hex.EncodeToString(sum)

	return name + "-" + suffix
}
