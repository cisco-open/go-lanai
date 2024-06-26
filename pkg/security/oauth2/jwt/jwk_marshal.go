package jwt

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
)

const (
	JwkTypeEC    = `EC`
	JwkTypeRSA   = `RSA`
	JwkTypeOctet = `oct`
	JwkTypeEdDSA = `OKP`
)

func marshalJwk(jwk Jwk) ([]byte, error) {
	params := generalJwk{Id: jwk.Id()}
	key := jwk.Public()
	var val interface{}
	switch v := key.(type) {
	case *rsa.PublicKey:
		val = makeRSAPublicJwk(v, params)
	case *ecdsa.PublicKey:
		val = makeECPublicJwk(v, params)
	case ed25519.PublicKey:
		val = makeOKPJwk(v, params)
	case []byte:
		val = makeOctetJwk(v, params)
	default:
		return nil, fmt.Errorf(`unable to marshal JWK: unrecognized public key type: %T`, key)
	}
	return json.Marshal(val)
}

func unmarshalJwk(data []byte) (Jwk, error) {
	var meta generalJwk
	if e := json.Unmarshal(data, &meta); e != nil {
		return nil, e
	}
	var jwk publicJwk
	switch meta.Type {
	case JwkTypeRSA:
		jwk = &rsaPublicJwk{}
	case JwkTypeEC:
		jwk = &ecPublicJwk{}
	case JwkTypeOctet:
		jwk = &octetJwk{}
	case JwkTypeEdDSA:
		jwk = &okpJwk{}
	default:
		return nil, fmt.Errorf(`invalid 'kty': %s`, meta.Type)
	}
	if e := json.Unmarshal(data, jwk); e != nil {
		return nil, e
	}
	return jwk.toJwk()
}

type jwkBytes []byte

func (b jwkBytes) String() string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func (b jwkBytes) BigInt() *big.Int {
	if len(b) == 0 {
		return nil
	}
	i := big.NewInt(0)
	i.SetBytes(b)
	return i
}

func (b jwkBytes) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *jwkBytes) UnmarshalText(data []byte) error {
	v, e := base64.RawURLEncoding.DecodeString(string(data))
	if e != nil {
		return e
	}
	*b = v
	return nil
}

type generalJwk struct {
	Id   string `json:"kid"`
	Type string `json:"kty"`
}

type publicJwk interface {
	toJwk() (Jwk, error)
}

// For EC key
// RFC 7518: https://datatracker.ietf.org/doc/html/rfc7518#section-6.2
type ecPublicJwk struct {
	generalJwk
	Curve       string   `json:"crv"`
	CoordinateX jwkBytes `json:"x"`
	CoordinateY jwkBytes `json:"y,omitempty"`
}

func (j ecPublicJwk) toJwk() (Jwk, error) {
	var curve elliptic.Curve
	switch j.Curve {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf(`unsupported 'crv' of EC JWK`)
	}
	key := &ecdsa.PublicKey{
		Curve: curve,
		X:     j.CoordinateX.BigInt(),
		Y:     j.CoordinateY.BigInt(),
	}
	return NewJwk(j.Id, j.Id, key), nil
}

func makeECPublicJwk(key *ecdsa.PublicKey, params generalJwk) ecPublicJwk {
	var x, y []byte
	if key.X != nil {
		x = key.X.Bytes()
	}
	if key.Y != nil {
		y = key.Y.Bytes()
	}
	var crv string
	if key.Curve.Params() != nil {
		crv = key.Curve.Params().Name
	}
	params.Type = JwkTypeEC
	return ecPublicJwk{
		generalJwk:  params,
		Curve:       crv,
		CoordinateX: x,
		CoordinateY: y,
	}
}

// For RSA key
// RFC 7518: https://datatracker.ietf.org/doc/html/rfc7518#section-6.3
type rsaPublicJwk struct {
	generalJwk
	Modulus  jwkBytes `json:"n"`
	Exponent jwkBytes `json:"e"`
}

func (j rsaPublicJwk) toJwk() (Jwk, error) {
	key := &rsa.PublicKey{
		N: j.Modulus.BigInt(),
		E: int(j.Exponent.BigInt().Uint64()),
	}
	return NewJwk(j.Id, j.Id, key), nil
}

func makeRSAPublicJwk(key *rsa.PublicKey, params generalJwk) rsaPublicJwk {
	params.Type = JwkTypeRSA
	return rsaPublicJwk{
		generalJwk: params,
		Modulus:    key.N.Bytes(),
		// Exponent convert to two's-complement in big-endian byte-order
		Exponent: bigEndian(key.E),
	}
}

// For symmetric key
// RFC 7518: https://datatracker.ietf.org/doc/html/rfc7518#section-6.4
type octetJwk struct {
	generalJwk
	Key jwkBytes `json:"k"`
}

func (j octetJwk) toJwk() (Jwk, error) {
	return NewJwk(j.Id, j.Id, []byte(j.Key)), nil
}

func makeOctetJwk(key []byte, params generalJwk) octetJwk {
	params.Type = JwkTypeOctet
	return octetJwk{
		generalJwk: params,
		Key:        key,
	}
}

// For ed25519 key. "OKP" (Octet Public Pair)
// RFC 8037: https://datatracker.ietf.org/doc/html/rfc8037#section-2
type okpJwk struct {
	generalJwk
	Curve     string   `json:"crv"`
	PublicKey jwkBytes `json:"x"`
}

func (j okpJwk) toJwk() (Jwk, error) {
	return NewJwk(j.Id, j.Id, ed25519.PublicKey(j.PublicKey)), nil
}

func makeOKPJwk(key ed25519.PublicKey, params generalJwk) okpJwk {
	params.Type = JwkTypeEdDSA
	return okpJwk{
		generalJwk: params,
		Curve:      "Ed25519",
		PublicKey:  jwkBytes(key),
	}
}

func bigEndian(i int) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	if e := binary.Write(buf, binary.BigEndian, uint64(i)); e != nil {
		return nil
	}
	// remove leading zeros
	data := buf.Bytes()
	for j := range data {
		if data[j] != 0 {
			data = data[j:]
			break
		}
	}
	return data
}
