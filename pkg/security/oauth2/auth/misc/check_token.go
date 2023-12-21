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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/claims"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
)

const (
	msgInvalidTokenType = "unsupported token type"
	msgInvalidToken     = "token is invalid or expired"
	hintAccessToken     = "access_token"
	hintRefreshToken    = "refresh_token"
)

type CheckTokenRequest struct {
	Token     string `form:"token"`
	Hint      string `form:"token_type_hint"`
	NoDetails bool   `form:"no_details"`
}

type CheckTokenEndpoint struct {
	issuer           security.Issuer
	authenticator    security.Authenticator
	tokenStoreReader oauth2.TokenStoreReader
}

func NewCheckTokenEndpoint(issuer security.Issuer, tokenStoreReader oauth2.TokenStoreReader) *CheckTokenEndpoint {
	authenticator := tokenauth.NewAuthenticator(func(opt *tokenauth.AuthenticatorOption) {
		opt.TokenStoreReader = tokenStoreReader
	})
	return &CheckTokenEndpoint{
		issuer:           issuer,
		authenticator:    authenticator,
		tokenStoreReader: tokenStoreReader,
	}
}

// CheckToken is the token introspection end point as defined in https://datatracker.ietf.org/doc/html/rfc7662
// This endpoint is used by protected resources to query the authorization server to determine the state and metadata of a token.
// Because this request is issued by a protected resource, the client used by the protected resource is not going to be the same
// as the client the token is issued for.
// The auth server require the protected resource to be specifically authorized to call this endpoint by means of client authentication
// and client scope (token_details).
// This end point is not meant to be used for other means. Any client that's not a protected resource should not be given this scope.
func (ep *CheckTokenEndpoint) CheckToken(c context.Context, request *CheckTokenRequest) (response *CheckTokenClaims, err error) {
	client := auth.RetrieveAuthenticatedClient(c)
	if client == nil {
		return nil, oauth2.NewInvalidClientError("check token endpoint requires client authentication")
	}

	switch request.Hint {
	case "":
		fallthrough
	case hintAccessToken:
		if request.NoDetails || !ep.allowDetails(c, client) {
			return ep.checkAccessTokenWithoutDetails(c, request)
		}
		return ep.checkAccessTokenWithDetails(c, request)
	case hintRefreshToken:
		return ep.checkRefreshToken(c, request)
	default:
		return nil, oauth2.NewUnsupportedTokenTypeError(fmt.Sprintf("token_type_hint '%s' is not supported", request.Hint))
	}
}

func (ep *CheckTokenEndpoint) allowDetails(_ context.Context, client oauth2.OAuth2Client) bool {
	return client.Scopes() != nil && client.Scopes().Has(oauth2.ScopeTokenDetails)
}

func (ep *CheckTokenEndpoint) checkAccessTokenWithoutDetails(c context.Context, request *CheckTokenRequest) (response *CheckTokenClaims, err error) {
	token, e := ep.tokenStoreReader.ReadAccessToken(c, request.Token)
	if e != nil || token == nil || token.Expired() {
		//nolint:nilerr // we hide error in response and returns compliant response
		return ep.inactiveTokenResponse(), nil
	}
	return ep.activeTokenResponseWithoutDetails(), nil
}

func (ep *CheckTokenEndpoint) checkAccessTokenWithDetails(c context.Context, request *CheckTokenRequest) (response *CheckTokenClaims, err error) {
	candidate := tokenauth.BearerToken{
		Token:      request.Token,
		DetailsMap: map[string]interface{}{},
	}
	oauth, e := ep.authenticator.Authenticate(c, &candidate)
	if e != nil || oauth.State() < security.StateAuthenticated {
		//nolint:nilerr // we hide error in response and returns compliant response
		return ep.inactiveTokenResponse(), nil
	}

	return ep.activeTokenResponseWithDetails(c, oauth.(oauth2.Authentication)), nil
}

func (ep *CheckTokenEndpoint) checkRefreshToken(_ context.Context, request *CheckTokenRequest) (response *CheckTokenClaims, err error) {
	// We don't support refresh token check for now
	return nil, oauth2.NewUnsupportedTokenTypeError(fmt.Sprintf("token_type_hint '%s' is not supported", request.Hint))
}

func (ep *CheckTokenEndpoint) inactiveTokenResponse() *CheckTokenClaims {
	return &CheckTokenClaims{
		Active: &utils.FALSE,
	}
}

func (ep *CheckTokenEndpoint) activeTokenResponseWithoutDetails() *CheckTokenClaims {
	return &CheckTokenClaims{
		Active: &utils.TRUE,
	}
}

func (ep *CheckTokenEndpoint) activeTokenResponseWithDetails(ctx context.Context, auth oauth2.Authentication) *CheckTokenClaims {
	c := CheckTokenClaims{
		Active: &utils.TRUE,
	}

	e := claims.Populate(ctx, &c,
		claims.WithSpecs(claims.CheckTokenClaimSpecs),
		claims.WithSource(auth),
		claims.WithIssuer(ep.issuer))

	if e != nil {
		return ep.activeTokenResponseWithoutDetails()
	}

	return &c
}

// Old impl. without claims factory, for reference only
//func (ep *CheckTokenEndpoint) activeTokenResponseWithDetails(auth oauth2.Authentication) *CheckTokenClaims {
//	claims := CheckTokenClaims{
//		Active: &utils.TRUE,
//		BasicClaims: oauth2.BasicClaims{
//			Audience:  auth.OAuth2Request().ClientId(),
//			ExpiresAt: auth.Details().(security.ContextDetails).ExpiryTime(),
//			//Id: auth.AccessToken().Id,
//			IssuedAt: auth.Details().(security.ContextDetails).IssueTime(),
//			//Issuer: auth.AccessToken(),
//			//NotBefore: auth.AccessToken(),
//			Subject:  auth.UserAuthentication().Principal().(string),
//			Scopes:   auth.OAuth2Request().Scopes(),
//			ClientId: auth.OAuth2Request().ClientId(),
//		},
//		Username:  auth.UserAuthentication().Principal().(string),
//		AuthTime:  auth.Details().(security.ContextDetails).AuthenticationTime(),
//		FirstName: auth.Details().(security.UserDetails).FirstName(),
//		LastName:  auth.Details().(security.UserDetails).LastName(),
//		Email:     auth.Details().(security.UserDetails).Email(),
//		Locale:    auth.Details().(security.UserDetails).LocaleCode(),
//
//		UserId:          auth.Details().(security.UserDetails).UserId(),
//		AccountType:     auth.Details().(security.UserDetails).AccountType().String(),
//		Currency:        auth.Details().(security.UserDetails).CurrencyCode(),
//		AssignedTenants: auth.Details().(security.UserDetails).AssignedTenantIds(),
//		TenantId:        auth.Details().(security.TenantDetails).TenantId(),
//		TenantExternalId:      auth.Details().(security.TenantDetails).TenantExternalId(),
//		TenantSuspended: utils.BoolPtr(auth.Details().(security.TenantDetails).TenantSuspended()),
//		ProviderId:      auth.Details().(security.ProviderDetails).ProviderId(),
//		ProviderName:    auth.Details().(security.ProviderDetails).ProviderName(),
//		Roles:           auth.Details().(security.ContextDetails).Roles(),
//		Permissions:     auth.Details().(security.ContextDetails).Permissions(),
//		OrigUsername:    auth.Details().(security.ProxiedUserDetails).OriginalUsername(),
//	}
//
//	return &claims
//}
