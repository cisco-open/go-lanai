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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
    "github.com/cisco-open/go-lanai/pkg/security/oauth2/tokenauth"
    "github.com/cisco-open/go-lanai/pkg/utils"
)

var (
	permissionBasedIgnoreParams = utils.NewStringSet(
		oauth2.ParameterClientSecret,
		oauth2.ParameterAccessToken,
	)
)

// PermissionBasedGranter is a helper based struct that provide common permission based implementations
type PermissionBasedGranter struct {
	authenticator security.Authenticator
}

func (g *PermissionBasedGranter) authenticateToken(ctx context.Context, request *auth.TokenRequest) (oauth2.Authentication, error) {
	tokenValue, ok := request.Extensions[oauth2.ParameterAccessToken].(string)
	if !ok {
		return nil, oauth2.NewInvalidTokenRequestError("access_token is missing")
	}

	candidate := tokenauth.BearerToken{
		Token:      tokenValue,
		DetailsMap: map[string]interface{}{},
	}

	// Authenticate
	auth, e := g.authenticator.Authenticate(ctx, &candidate)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e)
	}
	oauth, ok := auth.(oauth2.Authentication)
	switch {
	case !ok:
		fallthrough
	case oauth.State() < security.StateAuthenticated:
		return nil, oauth2.NewInvalidGrantError("invalid access token", e)
	case oauth.UserAuthentication() == nil || oauth.UserAuthentication().State() < security.StateAuthenticated:
		return nil, oauth2.NewInvalidGrantError("access token is not associated with a valid user")
	}

	return oauth, nil
}

func (g *PermissionBasedGranter) validateStoredPermissions(ctx context.Context, stored security.Authentication, permissions ...string) error {
	perms := stored.Permissions()
	if perms == nil {
		return oauth2.NewInvalidGrantError("user has no permissions")
	}
	for _, p := range permissions {
		if _, ok := perms[p]; !ok {
			return oauth2.NewInvalidGrantError(fmt.Sprintf("user doesn't have required permission [%s]", p))
		}
	}
	return nil
}

// Expectation is that only users with appropriate VIEW_OPERATOR_LOGIN_AS_CUSTOMER and
// SWITCH_TENANT permissions along with appropriate grant type are allowed to perform the security context
// switch.  This enforcement is done in other parts of the security context switch flow.
func (g *PermissionBasedGranter) validateStoredClient(ctx context.Context, client oauth2.OAuth2Client, src oauth2.OAuth2Request) error {
	original := src.ClientId()
	requested := client.ClientId()

	if original != requested {
		return oauth2.NewInvalidGrantError(fmt.Sprintf("security context switch as original Client ID [%s] and requesting Client ID [%s]", original, requested))
	}
	return nil
}

// Since we don't require requesting clientId to be same as original clientId, we have to also check
// original scope and requested scope. Ideally, when requesting clientId is different from original clientId,
// scopes should be re-authorized by user if it changed. However, since we always uses auto-approve,
// we could skip this step as long as all requested scope are auto approve.
//
// New scopes should be copied from either original request (if no "scope" param) or the token request.
// in both cases, they need to be validated against current client
func (g *PermissionBasedGranter) reduceScope(ctx context.Context, client oauth2.OAuth2Client,
	src oauth2.OAuth2Request, request *auth.TokenRequest) (oauth2.OAuth2Request, error) {

	original := src.Scopes()
	scopes := request.Scopes
	if scopes == nil || len(scopes) == 0 {
		scopes = original
	}

	if client.ClientId() != src.ClientId() {
		// we are dealing with different client, all scopes need to be re-validated against current client
		if e := auth.ValidateAllScopes(ctx, client, scopes); e != nil {
			return nil, e
		}
		if e := auth.ValidateAllAutoApprovalScopes(ctx, client, scopes); e != nil {
			return nil, e
		}
	} else {
		// same client, we only check if 1. new scope is a subset of original, OR 2. all new scopes are auto approved
		for scope, _ := range scopes {
			if !original.Has(scope) && !client.AutoApproveScopes().Has(scope) {
				return nil, oauth2.NewInvalidScopeError(fmt.Sprintf("scope [%s] is not allowed by this client", scope))
			}
		}
	}

	return src.NewOAuth2Request(func(opt *oauth2.RequestDetails) {
		opt.ClientId = client.ClientId()
		opt.RedirectUri = ""
		opt.GrantType = request.GrantType
		opt.Scopes = scopes
		for k, v := range request.Parameters {
			if permissionBasedIgnoreParams.Has(k) {
				continue
			}
			opt.Parameters[k] = v
		}
		for k, v := range request.Extensions {
			if permissionBasedIgnoreParams.Has(k) {
				continue
			}
			opt.Extensions[k] = v
		}
	}), nil
}
