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
	LoadByKid(ctx context.Context, kid string) (Jwk, error)
	LoadByName(ctx context.Context, name string) (Jwk, error)
	LoadAll(ctx context.Context) ([]Jwk, error)
}

type JwkRotator interface {
	JwkStore
	Rotate(ctx context.Context) error
	Current(ctx context.Context) (Jwk, error)
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


