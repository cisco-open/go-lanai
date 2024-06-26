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
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/jwt"
	"github.com/cisco-open/go-lanai/pkg/web"
	"net/http"
)

const (
	JwkTypeRSA = "RSA"
)

type JwkSetRequest struct {
	Kid string `uri:"kid"`
}

type JwkSetResponse struct {
	Keys []jwt.Jwk `json:"keys"`
}

type JwkSetEndpoint struct {
	jwkStore jwt.JwkStore
}

func NewJwkSetEndpoint(jwkStore jwt.JwkStore) *JwkSetEndpoint {
	return &JwkSetEndpoint{
		jwkStore: jwkStore,
	}
}

func (ep *JwkSetEndpoint) JwkByKid(ctx context.Context, req *JwkSetRequest) (resp jwt.Jwk, err error) {
	jwk, e := ep.jwkStore.LoadByKid(ctx, req.Kid)
	if e != nil {
		return nil, web.NewHttpError(http.StatusNotFound, e)
	}
	return jwk, nil
}

func (ep *JwkSetEndpoint) JwkSet(ctx context.Context, _ *JwkSetRequest) (resp *JwkSetResponse, err error) {
	jwks, e := ep.jwkStore.LoadAll(ctx)
	if e != nil {
		return nil, oauth2.NewGenericError(e.Error())
	}

	resp = &JwkSetResponse{
		Keys: jwks,
	}
	return resp, nil
}

