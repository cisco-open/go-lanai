// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package misc

import (
    "bytes"
    "context"
    "crypto/rsa"
    "encoding/base64"
    "encoding/binary"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
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

func (ep *JwkSetEndpoint) JwkByKid(c context.Context, _ *JwkSetRequest) (response *JwkSetResponse, err error) {
	return nil, fmt.Errorf("not implemented")
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
