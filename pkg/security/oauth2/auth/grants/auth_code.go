package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"fmt"
)

// ClientCredentialsGranter implements auth.TokenGranter
type AuthorizationCodeGranter struct {
	authService auth.AuthorizationService
}

func NewAuthorizationCodeGranter(authService auth.AuthorizationService) *AuthorizationCodeGranter {
	if authService == nil {
		panic(fmt.Errorf("cannot create AuthorizationCodeGranter without token service."))
	}

	return &AuthorizationCodeGranter{
		authService: authService,
	}
}

func (g *AuthorizationCodeGranter) Grant(ctx context.Context, request *auth.TokenRequest) (oauth2.AccessToken, error) {
	if oauth2.GrantTypeAuthCode != request.GrantType {
		return nil, nil
	}

	client := auth.RetrieveAuthenticatedClient(ctx)

	// common check
	if e := CommonPreGrantValidation(ctx, client, request); e != nil {
		return nil, e
	}

	// TODO create real authentication
	// create authentication
	req := request.OAuth2Request(client)
	oauth, e := g.authService.CreateAuthentication(ctx, req, nil)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e.Error(), e)
	}

	// create token
	token, e := g.authService.CreateAccessToken(ctx, oauth)
	if e != nil {
		return nil, oauth2.NewInvalidGrantError(e.Error(), e)
	}
	return token, nil
}


