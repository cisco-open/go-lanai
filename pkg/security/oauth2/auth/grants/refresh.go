package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
)

var (
	refreshIgnoreParams = utils.NewStringSet(
		oauth2.ParameterClientSecret,
		oauth2.ParameterRefreshToken,
	)
)

// RefreshGranter implements auth.TokenGranter
type RefreshGranter struct {
	authService auth.AuthorizationService
	tokenStore  auth.TokenStore
}

func NewRefreshGranter(authService auth.AuthorizationService, tokenStore  auth.TokenStore) *RefreshGranter {
	if authService == nil {
		panic(fmt.Errorf("cannot create AuthorizationCodeGranter without auth service."))
	}

	return &RefreshGranter{
		authService: authService,
		tokenStore:  tokenStore,
	}
}

func (g *RefreshGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if oauth2.GrantTypeRefresh != request.GrantType {
		return nil, nil
	}

	// for refresh grant, client have to be authenticated via client/secret
	client := auth.RetrieveAuthenticatedClient(ctx)
	if client == nil {
		return nil, oauth2.NewInvalidGrantError("client_credentials requires client secret validated")
	}

	// common check
	if e := CommonPreGrantValidation(ctx, client, request); e != nil {
		return nil, e
	}

	// extract refresh token
	refresh, ok := request.Extensions[oauth2.ParameterRefreshToken].(string)
	if !ok || refresh == "" {
		return nil, oauth2.NewInvalidTokenRequestError(fmt.Sprintf("missing required parameter %s", oauth2.ParameterRefreshToken))
	}

	refreshToken, e := g.tokenStore.ReadRefreshToken(ctx, refresh)
	if e != nil  {
		return nil, oauth2.NewInvalidGrantError(e)
	} else if refreshToken.WillExpire() && refreshToken.Expired() {
		_ = g.tokenStore.RemoveRefreshToken(ctx, refreshToken)
		return nil, oauth2.NewInvalidGrantError("refresh token expired")
	}

	// load stored authentication
	stored, e := g.tokenStore.ReadAuthentication(ctx, refresh, oauth2.TokenHintRefreshToken)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e)
	}

	// validate stored authentication
	// check client ID
	if stored.OAuth2Request().ClientId() != client.ClientId() {
		return nil, oauth2.NewInvalidGrantError("client ID mismatch")
	}

	// reduced scope
	oauthRequest, e := reduceScope(ctx, client, stored.OAuth2Request(), request)
	if e != nil {
		return nil, e
	}

	// construct auth
	// Note: user's authentication/details should be reloaded and re-verified in this process.
	oauth, e := g.authService.CreateAuthentication(ctx, oauthRequest, stored.UserAuthentication())
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e)
	}

	// create token
	token, e := g.authService.RefreshAccessToken(ctx, oauth, refreshToken)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e)
	}
	return token, nil
}

func reduceScope(c context.Context, client oauth2.OAuth2Client, src oauth2.OAuth2Request, request *auth.TokenRequest) (oauth2.OAuth2Request, error) {
	if !src.Approved() {
		return nil, oauth2.NewInvalidGrantError("original OAuth2 request was not approved")
	}

	if request.Scopes == nil || len(request.Scopes) == 0 {
		// didn't request scope reduction. bail
		return src, nil
	}

	// we double check if all requested scopes are authorized
	if e := auth.ValidateAllScopes(c, client, request.Scopes); e != nil {
		return nil, e
	}

	if ok, invalid := auth.IsSubSet(c, src.Scopes(), request.Scopes); !ok {
		return nil, oauth2.NewInvalidScopeError(fmt.Sprintf("scope [%s] was not originally authorized", invalid))
	}

	return src.NewOAuth2Request(func(opt *oauth2.RequestDetails) {
		opt.GrantType = request.GrantType
		opt.Scopes = request.Scopes
		for k, v := range request.Parameters {
			if refreshIgnoreParams.Has(k) {
				continue
			}
			opt.Parameters[k] = v
		}
		for k, v := range request.Extensions {
			if refreshIgnoreParams.Has(k) {
				continue
			}
			opt.Extensions[k] = v
		}
	}), nil
}

