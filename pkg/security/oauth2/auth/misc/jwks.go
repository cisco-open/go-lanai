package misc

import (
	"bytes"
	"context"
	"crypto/rsa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"encoding/base64"
	"encoding/binary"
)

const (
	JwkTypeRSA = "RSA"
)

type JwkSetRequest struct {

}

type JwkSetResponse struct {
	Keys []*JwkResponse `json:"keys"`
}

type JwkResponse struct {
	Id       string `json:"kid"`
	Type     string `json:"kty"`
	Modulus  string `json:"n"`
	Exponent string `json:"e"`
}

type JwkSetEndpoint struct {
	jwkStore jwt.JwkStore
}

func NewJwkSetEndpoint(jwkStore jwt.JwkStore) *JwkSetEndpoint {
	return &JwkSetEndpoint{
		jwkStore: jwkStore,
	}
}

func (ep *JwkSetEndpoint) JwkSet(c context.Context, _ *JwkSetRequest) (response *JwkSetResponse, err error) {
	jwks, e := ep.jwkStore.LoadAll(c)
	if e != nil {
		return nil, oauth2.NewGenericError(e.Error())
	}

	resp := JwkSetResponse{
		Keys: []*JwkResponse{},
	}
	for _, jwk := range jwks {
		if _, ok := jwk.Public().(*rsa.PublicKey); !ok {
			continue
		}
		jwkResp := convertRSA(jwk)
		if jwkResp != nil {
			resp.Keys = append(resp.Keys, jwkResp)
		}
	}
	return &resp, nil
}

func convertRSA(jwk jwt.Jwk) *JwkResponse {
	pubkey := jwk.Public().(*rsa.PublicKey)

	N := base64.RawURLEncoding.EncodeToString(pubkey.N.Bytes())

	// Exponent convert to two's-complement in big-endian byte-order
	buf := bytes.NewBuffer([]byte{})
	if e := binary.Write(buf, binary.BigEndian, int64(pubkey.E)); e != nil {
		return nil
	}
	E := base64.RawURLEncoding.EncodeToString(buf.Bytes())

	return &JwkResponse{
		Id:       jwk.Id(),
		Type:     JwkTypeRSA,
		Modulus:  N,
		Exponent: E,
	}
}
