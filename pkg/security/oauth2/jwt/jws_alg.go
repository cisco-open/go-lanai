package jwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
)

func resolveSigningMethod(key crypto.PrivateKey) (jwt.SigningMethod, error) {
	switch v := key.(type) {
	case *rsa.PrivateKey:
		return jwt.SigningMethodRS256, nil
	case *ecdsa.PrivateKey:
		size := v.Curve.Params().BitSize
		switch {
		case size >= 521:
			return jwt.SigningMethodES512, nil
		case size >= 384:
			return jwt.SigningMethodES384, nil
		case size >= 256:
			return jwt.SigningMethodES256, nil
		default:
			return nil, fmt.Errorf(`invalid ECDSA private key. Expect P-256 or more, but got %s`, v.Curve.Params().Name)
		}
	case ed25519.PrivateKey:
		return jwt.SigningMethodEdDSA, nil
	case []byte:
		switch {
		case len(v) >= 512/8:
			return jwt.SigningMethodHS512, nil
		case len(v) >= 384/8:
			return jwt.SigningMethodHS384, nil
		case len(v) >= 256/8:
			return jwt.SigningMethodHS256, nil
		default:
			return nil, fmt.Errorf(`invalid MAC secret. Expect 256B or more, but got %dB`, len(v))
		}
	default:
		return nil, fmt.Errorf(`unable to find proper signing method: unrecognized private key type: %T`, key)
	}
}

func generateCompatiblePrivateKey(method jwt.SigningMethod) (crypto.PrivateKey, error) {
	switch method {
	case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512:
		return rsa.GenerateKey(rand.Reader, rsaKeySize)
	case jwt.SigningMethodES256:
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case jwt.SigningMethodES384:
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case jwt.SigningMethodES512:
		return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	case jwt.SigningMethodPS256, jwt.SigningMethodPS384, jwt.SigningMethodPS512:
		return rsa.GenerateKey(rand.Reader, rsaKeySize)
	case jwt.SigningMethodEdDSA:
		_, priv, e := ed25519.GenerateKey(rand.Reader)
		return priv, e
	case jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512:
		// RFC7518: When using HMAC
		// "key of the same size as the hash output (for instance, 256 bits for "HS256") or larger MUST be used with this algorithm."
		var secret []byte
		switch method {
		case jwt.SigningMethodHS256:
			secret = make([]byte, 256/8)
		case jwt.SigningMethodHS384:
			secret = make([]byte, 384/8)
		case jwt.SigningMethodHS512:
			secret = make([]byte, 512/8)
		}
		if _, e := rand.Reader.Read(secret); e != nil {
			return nil, e
		}
		return secret, nil
	default:
		return nil, fmt.Errorf(`unsupported signing method: %T`, method)
	}
}

func generateRandomJwk(method jwt.SigningMethod, kid, name string) (Jwk, error) {
	privKey, e := generateCompatiblePrivateKey(method)
	if e != nil {
		return nil, e
	}
	return NewPrivateJwk(kid, name, privKey), nil
}
