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

package auth

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/google/uuid"
)

var (
	refreshTokenAllowedGrants = utils.NewStringSet(
		oauth2.GrantTypeAuthCode,
		oauth2.GrantTypeImplicit,
		oauth2.GrantTypeRefresh,
		oauth2.GrantTypeSwitchTenant, // Need this to create a new refresh token when switching tenants
		//oauth2.GrantTypePassword, // this is for dev purpose, shouldn't be allowed
	)
)

/*****************************
	RefreshToken Enhancer
 *****************************/

// RefreshTokenEnhancer implements order.Ordered and TokenEnhancer
// RefreshTokenEnhancer is responsible to create refresh token and associate it with the given access token
type RefreshTokenEnhancer struct {
	tokenStore TokenStore
	issuer     security.Issuer
}

func (te *RefreshTokenEnhancer) Order() int {
	return TokenEnhancerOrderRefreshToken
}

func (te *RefreshTokenEnhancer) Enhance(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication) (oauth2.AccessToken, error) {
	// step 1, check if refresh token is needed
	client, ok := ctx.Value(oauth2.CtxKeyAuthenticatedClient).(oauth2.OAuth2Client)
	if !ok || !te.isRefreshTokenNeeded(ctx, token, oauth, client) {
		return token, nil
	}

	t, ok := token.(*oauth2.DefaultAccessToken)
	if !ok {
		return nil, oauth2.NewInternalError("unsupported token implementation %T", t)
	}

	// step 2, create refresh token
	// Note: we don't reuse refresh token
	id := uuid.New().String()
	refresh := oauth2.NewDefaultRefreshToken(id)

	// step 3, set expriy time
	// Note: refresh token's validity is counted since authentication time
	details, ok := oauth.Details().(security.AuthenticationDetails)
	if ok && client.RefreshTokenValidity() > 0 && !details.AuthenticationTime().IsZero() {
		expiry := details.AuthenticationTime().Add(client.RefreshTokenValidity())
		refresh.SetExpireTime(expiry)
	}

	// step 4 create claims,
	request := oauth.OAuth2Request()
	claims := oauth2.BasicClaims{
		Id:       id,
		Audience: oauth2.StringSetClaim(utils.NewStringSet(client.ClientId())),
		Issuer:   te.issuer.Identifier(),
		Scopes:   request.Scopes(),
	}

	if oauth.UserAuthentication() != nil {
		if sub, e := extractSubject(oauth.UserAuthentication()); e != nil {
			return nil, e
		} else {
			claims.Subject = sub
		}
	}

	if refresh.WillExpire() && !refresh.ExpiryTime().IsZero() {
		claims.Set(oauth2.ClaimExpire, refresh.ExpiryTime())
	}
	refresh.SetClaims(&claims)

	// step 5, save refresh token
	if saved, e := te.tokenStore.SaveRefreshToken(ctx, refresh, oauth); e == nil {
		t.SetRefreshToken(saved)
	}
	return t, nil
}

/*****************************
	Helpers
 *****************************/

func (te *RefreshTokenEnhancer) isRefreshTokenNeeded(ctx context.Context, token oauth2.AccessToken, oauth oauth2.Authentication, client oauth2.OAuth2Client) bool {
	// refresh grant should be allowed for the client
	if e := ValidateGrant(ctx, client, oauth2.GrantTypeRefresh); e != nil {
		return false
	}

	// only some grant types can return refresh token
	if !refreshTokenAllowedGrants.Has(oauth.OAuth2Request().GrantType()) {
		return false
	}

	// last, if given token already have an refresh token, no need to generate new
	return token.RefreshToken() == nil || token.RefreshToken().WillExpire() && token.RefreshToken().Expired()
}
