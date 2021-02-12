package jwt

import (
	"context"
	"crypto"
	"crypto/rsa"
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
	// LoadByKid returns the JWK associated with given name.
	// The method might return different JWK for same name, if the store is also support rotation
	// This method is usually used when encoding/encrypt JWT token
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
	Implements
 *********************/
// RsaKeyPair implements Jwk and PrivateJwk
type RsaKeyPair struct {
	kid string
	name string
	private *rsa.PrivateKey
}

func NewRsaPrivateJwk(kid string, name string, privateKey *rsa.PrivateKey) *RsaKeyPair {
	return &RsaKeyPair{kid: kid, name: name, private: privateKey}
}

func (k *RsaKeyPair) Id() string {
	return k.kid
}

func (k *RsaKeyPair) Name() string {
	return k.name
}

func (k *RsaKeyPair) Public() crypto.PublicKey {
	return k.private.Public()
}

func (k *RsaKeyPair) Private() crypto.PrivateKey {
	return k.private
}

// RsaPublicKey implements Jwk
type RsaPublicKey struct {
	kid string
	name string
	public *rsa.PublicKey
}

func NewRsaJwk(kid string, name string, publicKey *rsa.PublicKey) *RsaPublicKey {
	return &RsaPublicKey{kid: kid, name: name, public: publicKey}
}

func (k *RsaPublicKey) Id() string {
	return k.kid
}

func (k *RsaPublicKey) Name() string {
	return k.name
}

func (k *RsaPublicKey) Public() crypto.PublicKey {
	return k.public
}


