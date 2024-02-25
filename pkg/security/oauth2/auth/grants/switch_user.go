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
	securityinternal "github.com/cisco-open/go-lanai/internal/security"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
	"github.com/cisco-open/go-lanai/pkg/tenancy"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"strings"
)

var (
	switchUserPermissions = []string{security.SpecialPermissionSwitchUser}
)

// SwitchUserGranter implements auth.TokenGranter
type SwitchUserGranter struct {
	PermissionBasedGranter
	authService  auth.AuthorizationService
	accountStore security.AccountStore
}

func NewSwitchUserGranter(authService auth.AuthorizationService, authenticator security.Authenticator, accountStore security.AccountStore) *SwitchUserGranter {
	if authenticator == nil {
		panic(fmt.Errorf("cannot create SwitchUserGranter without authenticator."))
	}

	if authService == nil {
		panic(fmt.Errorf("cannot create SwitchUserGranter without authorization service."))
	}

	if accountStore == nil {
		panic(fmt.Errorf("cannot create SwitchUserGranter without account store."))
	}

	return &SwitchUserGranter{
		PermissionBasedGranter: PermissionBasedGranter{
			authenticator: authenticator,
		},
		authService:  authService,
		accountStore: accountStore,
	}
}

func (g *SwitchUserGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if oauth2.GrantTypeSwitchUser != request.GrantType {
		return nil, nil
	}

	client := auth.RetrieveAuthenticatedClient(ctx)

	// common check
	if e := auth.ValidateGrant(ctx, client, request.GrantType); e != nil {
		return nil, e
	}

	// additional request params check
	if e := g.validateRequest(ctx, request); e != nil {
		return nil, e
	}

	// extract existing auth
	stored, e := g.authenticateToken(ctx, request)
	if e != nil {
		return nil, e
	}

	// check permissions
	if e := g.validate(ctx, request, stored); e != nil {
		return nil, e
	}

	// additional check
	// check client details (if client ID matches, if all requested scope is allowed, etc)
	if e := g.validateStoredClient(ctx, client, stored.OAuth2Request()); e != nil {
		return nil, e
	}

	// get merged request with reduced scope
	req, e := g.reduceScope(ctx, client, stored.OAuth2Request(), request)
	if e != nil {
		return nil, e
	}

	// get user authentication
	userAuth, e := g.loadUserAuthentication(ctx, request)
	if e != nil {
		return nil, e
	}

	// create authentication
	oauth, e := g.authService.SwitchAuthentication(ctx, req, userAuth, stored)
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

func (g *SwitchUserGranter) validateRequest(ctx context.Context, request *auth.TokenRequest) error {
	// switch_username or switch_user_id need to present
	// if both available, we use username
	username, nameOk := request.Extensions[oauth2.ParameterSwitchUsername].(string)
	userId, idOk := request.Extensions[oauth2.ParameterSwitchUserId].(string)
	if !nameOk && !idOk {
		return oauth2.NewInvalidTokenRequestError(fmt.Sprintf("both [%s] and [%s] are missing", oauth2.ParameterSwitchUsername, oauth2.ParameterSwitchUserId))
	}

	if strings.TrimSpace(username) == "" && strings.TrimSpace(userId) == "" {
		return oauth2.NewInvalidTokenRequestError(fmt.Sprintf("both [%s] and [%s] are empty", oauth2.ParameterSwitchUsername, oauth2.ParameterSwitchUserId))
	}
	return nil
}

func (g *SwitchUserGranter) validate(ctx context.Context, request *auth.TokenRequest, stored security.Authentication) error {
	if e := g.PermissionBasedGranter.validateStoredPermissions(ctx, stored, switchUserPermissions...); e != nil {
		return e
	}

	srcTenants, ok := stored.Details().(securityinternal.TenantAccessDetails)
	if !ok {
		return oauth2.NewInvalidGrantError("access token is not associated with a list of tenants")
	}

	if !canAccessAllTenants(ctx, srcTenants.EffectiveAssignedTenantIds()) {
		return oauth2.NewInvalidGrantError("user needs to be able to access all tenants to switch user")
	}

	srcUser, ok := stored.Details().(security.UserDetails)
	if !ok {
		return oauth2.NewInvalidGrantError("access token is not associated with a valid user")
	}

	if proxy, ok := stored.Details().(security.ProxiedUserDetails); ok && proxy.Proxied() {
		return oauth2.NewInvalidGrantError("the access token represents a masqueraded context. Nested masquerading is not supported")
	}

	username, _ := request.Extensions[oauth2.ParameterSwitchUsername].(string)
	if strings.TrimSpace(username) == srcUser.Username() {
		return oauth2.NewInvalidGrantError("cannot switch to same user")
	}

	userId, _ := request.Extensions[oauth2.ParameterSwitchUserId].(string)
	if strings.TrimSpace(userId) == srcUser.UserId() {
		return oauth2.NewInvalidGrantError("cannot switch to same user")
	}
	return nil
}

func (g *SwitchUserGranter) loadUserAuthentication(ctx context.Context, request *auth.TokenRequest) (security.Authentication, error) {
	username, _ := request.Extensions[oauth2.ParameterSwitchUsername].(string)
	userId, _ := request.Extensions[oauth2.ParameterSwitchUserId].(string)

	username = strings.TrimSpace(username)
	userId = strings.TrimSpace(userId)
	var account security.Account
	var e error
	// we prefer username over userId
	switch {
	case username != "":
		if account, e = g.accountStore.LoadAccountByUsername(ctx, username); e != nil {
			return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("invalid %s [%s]", oauth2.ParameterSwitchUsername, username), e)
		}
	default:
		if account, e = g.accountStore.LoadAccountById(ctx, userId); e != nil {
			return nil, oauth2.NewInvalidGrantError(fmt.Sprintf("invalid %s [%s]", oauth2.ParameterSwitchUserId, userId), e)
		}
	}

	permissions := map[string]interface{}{}
	for _, v := range account.Permissions() {
		permissions[v] = true
	}

	return oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
		opt.Principal = account.Username()
		opt.Permissions = permissions
		opt.State = security.StateAuthenticated
	}), nil
}

func canAccessAllTenants(ctx context.Context, assignedTenantIds utils.StringSet) bool {
	if assignedTenantIds.Has(security.SpecialTenantIdWildcard) {
		return true
	}

	rootId, err := tenancy.GetRoot(ctx)
	if err != nil || rootId == "" {
		return false
	}
	return assignedTenantIds.Has(rootId)
}
