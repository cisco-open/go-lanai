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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"fmt"
)

// PasswordGranter implements auth.TokenGranter
type PasswordGranter struct {
	authenticator security.Authenticator
	authService   auth.AuthorizationService
}

func NewPasswordGranter(authService auth.AuthorizationService, authenticator security.Authenticator) *PasswordGranter {
	if authenticator == nil {
		panic(fmt.Errorf("cannot create PasswordGranter without authenticator."))
	}

	if authService == nil {
		panic(fmt.Errorf("cannot create PasswordGranter without authorization service."))
	}

	return &PasswordGranter{
		authenticator: authenticator,
		authService:   authService,
	}
}

func (g *PasswordGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if oauth2.GrantTypePassword != request.GrantType {
		return nil, nil
	}

	client := auth.RetrieveAuthenticatedClient(ctx)

	// common check
	if e := CommonPreGrantValidation(ctx, client, request); e != nil {
		return nil, e
	}

	// extract username & password
	username, uOk := request.Parameters[oauth2.ParameterUsername]
	password, pOk := request.Parameters[oauth2.ParameterPassword]
	delete(request.Parameters, oauth2.ParameterPassword)
	if !uOk || !pOk {
		return nil, oauth2.NewInvalidGrantError("missing 'username' and 'password'")
	}

	// authenticate
	candidate := passwd.UsernamePasswordPair{
		Username: username,
		Password: password,
	}

	userAuth, err := g.authenticator.Authenticate(ctx, &candidate)
	if err != nil || userAuth.State() < security.StateAuthenticated {
		return nil, oauth2.NewInvalidGrantError(err)
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
	oauth, e := g.authService.CreateAuthentication(ctx, req, userAuth)
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


