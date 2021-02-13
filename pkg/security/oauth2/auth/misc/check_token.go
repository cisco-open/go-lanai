package misc

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/tokenauth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
)

const (
	msgInvalidTokenType = "tnsupported token type"
	msgInvalidToken = "token is invalid or expired"
	hintAccessToken = "access_token"
	hintRefreshToken = "refresh_token"
)

type CheckTokenRequest struct {
	Token     string `form:"token"`
	Hint      string `form:"token_type_hint"`
	NoDetails bool   `form:"no_details"`
}

type CheckTokenEndpoint struct {
	authenticator    security.Authenticator
	tokenStoreReader oauth2.TokenStoreReader
}

func NewCheckTokenEndpoint(tokenStoreReader oauth2.TokenStoreReader) *CheckTokenEndpoint {
	authenticator := tokenauth.NewAuthenticator(func(opt *tokenauth.AuthenticatorOption) {
		opt.TokenStoreReader = tokenStoreReader
	})
	return &CheckTokenEndpoint{
		authenticator:    authenticator,
		tokenStoreReader: tokenStoreReader,
	}
}

func (ep *CheckTokenEndpoint) CheckToken(c context.Context, request *CheckTokenRequest) (response *CheckTokenClaims, err error) {
	// TODO check client auth

	switch request.Hint {
	case "":
		fallthrough
	case hintAccessToken:
		if request.NoDetails || !ep.allowDetails(c) {
			return ep.checkAccessTokenWithoutDetails(c, request)
		}
		return ep.checkAccessTokenWithDetails(c, request)
	case hintRefreshToken:
		return ep.checkRefreshToken(c, request)
	default:
		return nil, oauth2.NewUnsupportedTokenTypeError(fmt.Sprintf("token_type_hint '%s' is not supported", request.Hint))
	}
}

func (ep *CheckTokenEndpoint) allowDetails(c context.Context) bool {
	// TODO use client
	return true
}

func (ep *CheckTokenEndpoint) checkAccessTokenWithoutDetails(c context.Context, request *CheckTokenRequest) (response *CheckTokenClaims, err error) {
	token, e := ep.tokenStoreReader.ReadAccessToken(c, request.Token)
	if e != nil || token == nil || token.Expired() {
		return ep.inactiveTokenResponse(), nil
	}
	return ep.activeTokenResponseWithoutDetails(), nil
}

func (ep *CheckTokenEndpoint) checkAccessTokenWithDetails(c context.Context, request *CheckTokenRequest) (response *CheckTokenClaims, err error) {
	candidate := tokenauth.BearerToken{
		Token: request.Token,
		DetailsMap: map[string]interface{}{},
	}
	auth, e := ep.authenticator.Authenticate(c, &candidate)
	if e != nil || auth.State() < security.StateAuthenticated {
		return ep.inactiveTokenResponse(), nil
	}

	return ep.activeTokenResponseWithDetails(auth.(oauth2.Authentication)), nil
}

func (ep *CheckTokenEndpoint) checkRefreshToken(c context.Context, request *CheckTokenRequest) (response *CheckTokenClaims, err error) {
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

func (ep *CheckTokenEndpoint) activeTokenResponseWithDetails(auth oauth2.Authentication) *CheckTokenClaims {
	// TODO proper claims generation
	claims := CheckTokenClaims{
		Active: &utils.TRUE,
		BasicClaims: oauth2.BasicClaims{
			Audience:  auth.OAuth2Request().ClientId(),
			ExpiresAt: auth.Details().(security.ContextDetails).ExpiryTime(),
			//Id: auth.AccessToken().Id,
			IssuedAt: auth.Details().(security.ContextDetails).IssueTime(),
			//Issuer: auth.AccessToken(),
			//NotBefore: auth.AccessToken(),
			Subject:  auth.UserAuthentication().Principal().(string),
			Scopes:   auth.OAuth2Request().Scopes(),
			ClientId: auth.OAuth2Request().ClientId(),
		},
		Username:  auth.UserAuthentication().Principal().(string),
		AuthTime:  auth.Details().(security.ContextDetails).AuthenticationTime(),
		FirstName: auth.Details().(security.UserDetails).FirstName(),
		LastName:  auth.Details().(security.UserDetails).LastName(),
		Email:     auth.Details().(security.UserDetails).Email(),
		Locale:    auth.Details().(security.UserDetails).LocaleCode(),

		UserId:          auth.Details().(security.UserDetails).UserId(),
		AccountType:     auth.Details().(security.UserDetails).AccountType().String(),
		Currency:        auth.Details().(security.UserDetails).CurrencyCode(),
		AssignedTenants: auth.Details().(security.UserDetails).AssignedTenantIds(),
		TenantId:        auth.Details().(security.TenantDetails).TenantId(),
		TenantName:      auth.Details().(security.TenantDetails).TenantName(),
		TenantSuspended: utils.BoolPtr(auth.Details().(security.TenantDetails).TenantSuspended()),
		ProviderId:      auth.Details().(security.ProviderDetails).ProviderId(),
		ProviderName:    auth.Details().(security.ProviderDetails).ProviderName(),
		Roles:           auth.Details().(security.ContextDetails).Roles(),
		Permissions:     auth.Details().(security.ContextDetails).Permissions(),
		OrigUsername:    auth.Details().(security.ProxiedUserDetails).OriginalUsername(),
	}

	return &claims
}