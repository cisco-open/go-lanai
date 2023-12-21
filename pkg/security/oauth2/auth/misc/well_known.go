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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/openid"
	"net/http"
)

// WellKnownEndpoint provide "/.well-known/**" HTTP endpoints
type WellKnownEndpoint struct {
	issuer     security.Issuer
	extra      map[string]interface{}
}

func NewWellKnownEndpoint(issuer security.Issuer, idpManager idp.IdentityProviderManager, extra map[string]interface{}) *WellKnownEndpoint {
	if extra == nil {
		extra = map[string]interface{}{}
	}
	extra[openid.OPMetaExtraSourceIDPManager] = idpManager
	return &WellKnownEndpoint{
		issuer: issuer,
		extra:  extra,
	}
}

// OpenIDConfig should mapped to GET /.well-known/openid-configuration
func (ep *WellKnownEndpoint) OpenIDConfig(ctx context.Context, _ *http.Request) (resp *openid.OPMetadata, err error) {
	c := openid.OPMetadata{MapClaims: oauth2.MapClaims{}}
	e := claims.Populate(ctx, &c,
		claims.WithSpecs(openid.OPMetadataBasicSpecs, openid.OPMetadataOptionalSpecs),
		claims.WithIssuer(ep.issuer),
		claims.WithExtraSource(ep.extra),
	)
	if e != nil {
		return nil, e
	}
	return &c, nil
}
