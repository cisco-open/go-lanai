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

package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"fmt"
)

// ClientCredentialsGranter implements auth.TokenGranter
type ClientCredentialsGranter struct {
	authService auth.AuthorizationService
}

func NewClientCredentialsGranter(authService auth.AuthorizationService) *ClientCredentialsGranter {
	if authService == nil {
		panic(fmt.Errorf("cannot create ClientCredentialsGranter without token service."))
	}

	return &ClientCredentialsGranter{
		authService: authService,
	}
}

func (g *ClientCredentialsGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if oauth2.GrantTypeClientCredentials != request.GrantType {
		return nil, nil
	}

	// for client credentials grant, client have to be authenticated via client/secret
	client, e := auth.RetrieveFullyAuthenticatedClient(ctx)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError("client_credentials requires client secret validated")
	}

	// common check
	if e := CommonPreGrantValidation(ctx, client, request); e != nil {
		return nil, e
	}

	// additional check
	if request.Scopes == nil || len(request.Scopes) == 0 {
		request.Scopes = client.Scopes()
	}
	if e := auth.ValidateAllAutoApprovalScopes(ctx, client, request.Scopes); e != nil {
		return nil, e
	}

	// create authentication
	req := request.OAuth2Request(client)
	oauth, e := g.authService.CreateAuthentication(ctx, req, nil)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e)
	}

	// create token
	token, e := g.authService.CreateAccessToken(ctx, oauth)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e)
	}
	return token, nil
}

